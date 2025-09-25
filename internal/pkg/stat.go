package pkg

import (
    "io"
    "math"
    "os"
    "strconv"
    "sync"
    "sync/atomic"
    "time"

    "github.com/charmbracelet/bubbles/progress"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/valyala/bytebufferpool"
    "github.com/valyala/fasthttp"
)

const (
    done         = 1
    fieldWidth   = 18
    defaultFps   = time.Duration(40)
    padding      = 2
    maxWidth     = 66
    processColor = "#444"
)

type stat struct {
    r io.Reader
    w io.Writer

    throughput *int64
    reqs       int64
    elapsed    int64
    code1xx    int64
    code2xx    int64
    code3xx    int64
    code4xx    int64
    code5xx    int64
    codeOthers int64
    latencies  []int64
    rps        []float64
    mut        sync.Mutex
    errs       map[string]int
    buf        *bytebufferpool.ByteBuffer

    url         string
    count       int
    duration    time.Duration
    connections int
    initCmd     tea.Cmd
    progressBar progress.Model
    quitting    bool
    done        bool
}

func newStat() *stat {
    progressBar := progress.New(progress.WithSolidFill(processColor))

    return &stat{
        r:           os.Stdin,
        w:           os.Stdout,
        errs:        make(map[string]int),
        buf:         bytebufferpool.Get(),
        progressBar: progressBar,
    }
}

func (t *stat) start() error {
    return tea.NewProgram(t, tea.WithInput(t.r), tea.WithOutput(t.w)).Start()
}

func (t *stat) Init() tea.Cmd {
    return tea.Batch(tickNow, t.initCmd)
}

func (t *stat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            fallthrough
        case "esc":
            fallthrough
        case "ctrl+c":
            t.quitting = true
            return t, tea.Quit
        default:
            return t, nil
        }
    case tea.WindowSizeMsg:
        t.progressBar.Width = msg.Width - padding*2 - 4
        if t.progressBar.Width > maxWidth {
            t.progressBar.Width = maxWidth
        }
        return t, nil

    case int:
        var cmd tea.Cmd
        if msg == done {
            t.done = true
            cmd = tea.Quit
        }
        return t, cmd

    default:
        return t, tick()
    }

}

func (t *stat) View() string {
    return t.output()
}

func (t *stat) appendCode(code int) {
    switch code / 100 {
    case 1:
        t.code1xx++
    case 2:
        t.code2xx++
    case 3:
        t.code3xx++
    case 4:
        t.code4xx++
    case 5:
        t.code5xx++
    default:
        t.codeOthers++
    }
}

func (t *stat) appendRps(rps float64) {
    t.rps = append(t.rps, rps)
}

func (t *stat) appendLatency(latency time.Duration) {
    t.latencies = append(t.latencies, latency.Microseconds())
}

func (t *stat) appendError(err error) {
    t.mut.Lock()
    t.errs[err.Error()]++
    t.mut.Unlock()
}

func (t *stat) output() string {
    t.buf.Reset()

    t.writeTitle()
    t.writeProcessBar()
    t.writeTotalRequest()
    t.writeElapsed()
    t.writeThroughput()
    t.writeStatistics()
    t.writeCodes()
    t.writeErrors()
    t.writeHint()

    return t.buf.String()
}

func (t *stat) writeTitle() {
    _, _ = t.buf.WriteString("Benchmarking ")
    _, _ = t.buf.WriteString(t.url)
    _, _ = t.buf.WriteString(" with ")
    t.writeInt(t.connections)
    _, _ = t.buf.WriteString(" connections\n")
}

func (t *stat) writeProcessBar() {
    var percent float64
    if t.count != 0 {
        percent = float64(atomic.LoadInt64(&t.reqs)) / float64(t.count)
    } else {
        percent = float64(atomic.LoadInt64(&t.elapsed)) / float64(t.duration)
    }

    if percent > 1.0 {
        percent = 1.0
    }

    _, _ = t.buf.WriteString(t.progressBar.ViewAs(percent))
    _ = t.buf.WriteByte('\n')
}

func (t *stat) writeTotalRequest() {
    _, _ = t.buf.WriteString("Requests:  ")
    t.writeInt(int(atomic.LoadInt64(&t.reqs)))
    if t.count != 0 {
        _ = t.buf.WriteByte('/')
        t.writeInt(t.count)
    }
    _, _ = t.buf.WriteString("  ")
}

func (t *stat) writeElapsed() {
    elapsed := time.Duration(atomic.LoadInt64(&t.elapsed))
    _, _ = t.buf.WriteString("Elapsed:  ")
    if elapsed > t.duration {
        elapsed = t.duration
    }
    t.writeFloat(elapsed.Seconds())
    if t.count == 0 {
        _ = t.buf.WriteByte('/')
        t.writeFloat(t.duration.Seconds())
    }
    _, _ = t.buf.WriteString("s  ")
}

func (t *stat) writeThroughput() {
    _, _ = t.buf.WriteString("Throughput:  ")
    elapsed := time.Duration(atomic.LoadInt64(&t.elapsed))
    if seconds := elapsed.Seconds(); seconds != 0 {
        throughput, unit := formatThroughput(float64(atomic.LoadInt64(t.throughput)) / seconds)
        t.writeFloat(throughput)
        _ = t.buf.WriteByte(' ')
        _, _ = t.buf.WriteString(unit)
    } else {
        _, _ = t.buf.WriteString("0 B/s")
    }
    _ = t.buf.WriteByte('\n')
}

func (t *stat) writeStatistics() {
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(12).Align(lipgloss.Center).Render("Statistics  "))

    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(fieldWidth).Align(lipgloss.Center).Render("Avg"))
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(fieldWidth).Align(lipgloss.Center).Render("Stdev"))
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(fieldWidth).Align(lipgloss.Center).Render("Max"))
    _ = t.buf.WriteByte('\n')

    rpsAvg, rpsStdev, rpsMax := rpsResult(t.rps)
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(12).Align(lipgloss.Center).Render("Reqs/sec  "))

    t.writeRps(rpsAvg)
    t.writeRps(rpsStdev)
    t.writeRps(rpsMax)
    _ = t.buf.WriteByte('\n')

    latencyAvg, latencyStdev, latencyMax := latencyResult(t.latencies)
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(12).Align(lipgloss.Center).Render("Latency  "))
    t.writeLatency(latencyAvg)
    t.writeLatency(latencyStdev)
    t.writeLatency(latencyMax)
    _ = t.buf.WriteByte('\n')
}

func (t *stat) writeRps(rps float64) {
    s := strconv.FormatFloat(rps, 'f', 2, 64)
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(fieldWidth).Align(lipgloss.Center).Render(s))
}

func (t *stat) writeLatency(latency float64) {
    s := strconv.FormatFloat(latency, 'f', 2, 64)
    _, _ = t.buf.WriteString(lipgloss.NewStyle().Width(fieldWidth).Align(lipgloss.Center).Render(s + "ms"))
}

func (t *stat) writeCodes() {
    _, _ = t.buf.WriteString("HTTP codes:\n  ")

    _, _ = t.buf.WriteString("1xx - ")
    t.writeInt(int(atomic.LoadInt64(&t.code1xx)), "#ffaf00")
    _, _ = t.buf.WriteString(", ")

    _, _ = t.buf.WriteString("2xx - ")
    t.writeInt(int(atomic.LoadInt64(&t.code2xx)), "#00ff00")
    _, _ = t.buf.WriteString(", ")

    _, _ = t.buf.WriteString("3xx - ")
    t.writeInt(int(atomic.LoadInt64(&t.code3xx)), "#ffff00")
    _, _ = t.buf.WriteString(", ")

    _, _ = t.buf.WriteString("4xx - ")
    t.writeInt(int(atomic.LoadInt64(&t.code4xx)), "#ff8700")
    _, _ = t.buf.WriteString(", ")

    _, _ = t.buf.WriteString("5xx - ")
    t.writeInt(int(atomic.LoadInt64(&t.code5xx)), "#870000")
    _, _ = t.buf.WriteString("\n  ")

    _, _ = t.buf.WriteString("Others - ")
    t.writeInt(int(atomic.LoadInt64(&t.codeOthers)), "#444")
    _, _ = t.buf.WriteString("\n")
}

func (t *stat) writeErrors() {
    t.mut.Lock()
    defer t.mut.Unlock()

    if len(t.errs) == 0 {
        return
    }
    _, _ = t.buf.WriteString("Errors:\n")
    for err, count := range t.errs {
        _, _ = t.buf.WriteString("  ")
        _, _ = t.buf.WriteString(lipgloss.NewStyle().Underline(true).Render(err))
        _, _ = t.buf.WriteString(": ")
        t.writeInt(count)
        _ = t.buf.WriteByte('\n')
    }
}

func (t *stat) writeHint() {
    if t.done {
        _, _ = t.buf.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#008700")).Render(" Done! \n"))
    } else if t.quitting {
        _, _ = t.buf.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#870000")).Render(" Terminated! \n"))
    } else {
        _, _ = t.buf.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#444")).Render(" press q/esc/ctrl+c to quit "))
    }
}

func (t *stat) writeInt(i int, colorStr ...string) {
    if i <= 0 || len(colorStr) == 0 {
        t.buf.B = fasthttp.AppendUint(t.buf.B, i)
        return
    }

    _, _ = t.buf.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorStr[0])).Render(strconv.Itoa(i)))
}

func (t *stat) writeFloat(f float64) {
    t.buf.B = strconv.AppendFloat(t.buf.B, f, 'f', 2, 64)
}

func rpsResult(rps []float64) (avg float64, stdev float64, max float64) {
    l := len(rps)
    if l == 0 {
        return
    }

    var sum, sum2 float64
    for _, r := range rps {
        sum += r
        if r > max {
            max = r
        }
    }

    avg = sum / float64(l)

    var diff float64
    for _, r := range rps {
        diff = avg - r
        sum2 += diff * diff
    }

    stdev = math.Sqrt(sum2 / float64(l-1))

    return
}

func latencyResult(latencies []int64) (avg float64, stdev float64, max float64) {
    l := len(latencies)
    if l == 0 {
        return
    }

    var sum float64
    for _, latency := range latencies {
        // us -> ms
        r := float64(latency) / 1000
        sum += r
        if r > max {
            max = r
        }
    }

    avg = sum / float64(l)

    var diff, sum2 float64
    for _, latency := range latencies {
        r := float64(latency) / 1000
        diff = avg - r
        sum2 += diff * diff
    }

    stdev = math.Sqrt(sum2 / float64(l-1))

    return
}

func formatThroughput(throughput float64) (float64, string) {
    switch {
    case throughput < 1e3:
        return throughput, "B/s"
    case throughput < 1e6:
        return throughput / 1e3, "KB/s"
    case throughput < 1e9:
        return throughput / 1e6, "MB/s"
    default:
        return throughput / 1e9, "GB/s"
    }
}

type tickMsg struct {
    Time time.Time
}

func tickNow() tea.Msg {
    return tickMsg{Time: time.Now()}
}

func tick() tea.Cmd {
    return tea.Tick(time.Second/defaultFps, func(t time.Time) tea.Msg {
        return tickMsg{Time: t}
    })
}
