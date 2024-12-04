package main

import (
	"bufio"
	"context"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		printMenu()
		fmt.Print("请输入选项: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败，请重新输入")
			continue
		}
		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Println("无效的输入，请输入数字1-4")
			continue
		}

		switch choice {
		case 0:
			startProxy()
		case 1:
			modifyConfig(reader)
		case 2:
			viewCompareStats()
		case 3:
			viewLogs()
		case 4:
			fmt.Println("退出程序")
			return
		default:
			fmt.Println("无效选项，请输入数字0-4")
		}
	}
}

func printMenu() {
	fmt.Println("\n===== 主菜单 =====")
	fmt.Println("0. 启动程序")
	fmt.Println("1. 修改配置")
	fmt.Println("2. 查看当前 CompareStats")
	fmt.Println("3. 查看日志文件 diff.log 输出内容")
	fmt.Println("4. 退出程序")
}

func modifyConfig(reader *bufio.Reader) {
	for {
		fmt.Println("\n===== 配置修改菜单 =====")
		fmt.Println("0. 查看 当前所有配置")
		fmt.Println("1. 修改 SelfPort")
		fmt.Println("2. 修改 ProxyHost")
		fmt.Println("3. 修改 HeadersInclude")
		fmt.Println("4. 修改 CompareStatusCode")
		fmt.Println("5. 修改 EquivalentStatusCodes")
		fmt.Println("6. 修改 CompareBody")
		fmt.Println("7. 修改 BodiesExclude")
		fmt.Println("8. 返回主菜单")

		fmt.Print("请输入选项: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败，请重新输入")
			continue
		}
		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Println("无效的输入，请输入数字1-6")
			continue
		}

		switch choice {
		case 0:
			showConfig()
		case 1:
			modifySelfPort(reader)
		case 2:
			modifyProxyHost(reader)
		case 3:
			modifyHeadersInclude(reader)
		case 4:
			modifyCompareStatusCode(reader)
		case 5:
			modifyEquivalentStatusCodes(reader)
		case 6:
			modifyCompareBody(reader)
		case 7:
			modifyBodiesExclude(reader)
		case 8:
			return
		default:
			fmt.Println("无效选项，请输入数字0-8")
		}
	}
}

func startProxy() {
	startProxyAsync()
}

func showConfig() {
	fmt.Printf("config: %+v", config)
}

func modifySelfPort(reader *bufio.Reader) {
	fmt.Println("当前 SelfPort:", config.SelfPort)
	fmt.Print("请输入 SelfPort(:18080)：")
	input, _ := reader.ReadString('\n')
	config.SelfPort = strings.TrimSpace(input)
	fmt.Println("更新后的 SelfPort:", config.SelfPort)
}
func modifyProxyHost(reader *bufio.Reader) {
	fmt.Print("请输入 ProxyHost（格式：http://localhost:8081;http://localhost:8082）: ")
	input, _ := reader.ReadString('\n')
	entries := strings.Split(strings.TrimSpace(input), ";")
	if len(entries) != 2 {
		fmt.Println("无效的输入，请重试")
	}
	config.Host1 = entries[0]
	config.Host2 = entries[1]
	fmt.Printf("更新后的 Host1:%s, Host2:%s \n", config.Host1, config.Host2)
}
func modifyHeadersInclude(reader *bufio.Reader) {
	fmt.Println("当前 HeadersInclude:", config.HeadersInclude)
	fmt.Print("请输入要添加的 Headers（逗号分隔）：")
	input, _ := reader.ReadString('\n')
	headers := strings.Split(strings.TrimSpace(input), ",")
	config.HeadersInclude = append(config.HeadersInclude, headers...)
	fmt.Println("更新后的 HeadersInclude:", config.HeadersInclude)
}

func modifyCompareStatusCode(reader *bufio.Reader) {
	fmt.Print("请输入 CompareStatusCode (true/false): ")
	input, _ := reader.ReadString('\n')
	value := strings.TrimSpace(input)
	if value == "true" {
		config.CompareStatusCode = true
	} else if value == "false" {
		config.CompareStatusCode = false
	} else {
		fmt.Println("无效的输入，请输入 true 或 false")
		return
	}
	fmt.Println("更新后的 CompareStatusCode:", config.CompareStatusCode)
}

func modifyEquivalentStatusCodes(reader *bufio.Reader) {
	fmt.Print("请输入 EquivalentStatusCodes（格式：400,401;200,201）: ")
	input, _ := reader.ReadString('\n')
	entries := strings.Split(strings.TrimSpace(input), ";")
	var codes [][]int
	for _, entry := range entries {
		pair := strings.Split(entry, ",")
		if len(pair) != 2 {
			fmt.Println("无效的输入格式，请重试")
			return
		}
		code1, err1 := strconv.Atoi(pair[0])
		code2, err2 := strconv.Atoi(pair[1])
		if err1 != nil || err2 != nil {
			fmt.Println("无效的状态码，请重试")
			return
		}
		codes = append(codes, []int{code1, code2})
	}
	config.EquivalentStatusCodes = codes
	fmt.Println("更新后的 EquivalentStatusCodes:", config.EquivalentStatusCodes)
}

func modifyCompareBody(reader *bufio.Reader) {
	fmt.Print("请输入 CompareBody (true/false): ")
	input, _ := reader.ReadString('\n')
	value := strings.TrimSpace(input)
	if value == "true" {
		config.CompareBody = true
	} else if value == "false" {
		config.CompareBody = false
	} else {
		fmt.Println("无效的输入，请输入 true 或 false")
		return
	}
	fmt.Println("更新后的 CompareBody:", config.CompareBody)
}

func modifyBodiesExclude(reader *bufio.Reader) {
	fmt.Println("当前 BodiesExclude:", config.BodiesExclude)
	fmt.Print("请输入要添加的 Body Exclude 路径（逗号分隔）：")
	input, _ := reader.ReadString('\n')
	paths := strings.Split(strings.TrimSpace(input), ",")
	config.BodiesExclude = append(config.BodiesExclude, paths...)
	fmt.Println("更新后的 BodiesExclude:", config.BodiesExclude)
}
func viewCompareStats() {
	if err := ui.Init(); err != nil {
		logger.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	table := widgets.NewTable()
	table.Title = "Comparison Stats, Quit with Q"
	table.Rows = [][]string{{"Route", "Same Count", "Diff Count", "Example Header", "Example Status", "Example Body Key"}}
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.SetRect(0, 0, 120, 10)

	ui.Render(table)

	go func(ctx context.Context) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				compareStats := statsGroup.GetAllRouteStats()

				// 准备表格数据
				rows := [][]string{{"Route", "Same Count", "Diff Count", "Example Header", "Example Status", "Example Body Key"}}

				for _, stat := range compareStats {
					copyStat := stat.GetRouteStats()

					headerExample := ""
					for key, example := range copyStat.DifferenceDetails.HeaderDifferences {
						headerExample = fmt.Sprintf("%s: %s vs %s", key, example.Example1, example.Example2)
						break
					}

					statusExample := ""
					for key, example := range copyStat.DifferenceDetails.StatusDifferences {
						statusExample = fmt.Sprintf("%s: %s vs %s", key, example.Example1, example.Example2)
						break
					}

					bodyExample := ""
					for key, example := range copyStat.DifferenceDetails.BodyDifferences {
						bodyExample = fmt.Sprintf("%s: %s vs %s", key, example.Example1, example.Example2)
						break
					}

					rows = append(rows, []string{
						copyStat.Route,
						fmt.Sprintf("%d", copyStat.IdenticalCount),
						fmt.Sprintf("%d", copyStat.DifferenceCount),
						headerExample,
						statusExample,
						bodyExample,
					})
				}

				// 更新表格数据
				table.Rows = rows

				// 刷新视图
				ui.Render(table)
			}
		}
	}(ctx)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent && e.ID == "q" {
			cancel()
			break
		}
	}
}

func viewLogs() {
	if err := ui.Init(); err != nil {
		logger.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logParagraph := widgets.NewParagraph()
	logParagraph.Title = "Log Viewer (Press 'Q' to Quit)"
	logParagraph.Text = "Loading logs...\n"
	logParagraph.TextStyle = ui.NewStyle(ui.ColorWhite)
	logParagraph.SetRect(0, 0, 100, 24)
	ui.Render(logParagraph)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 打开日志文件
	file, err := os.Open(logFile)
	defer file.Close()

	if err != nil {
		logParagraph.Text = fmt.Sprintf("无法打开日志文件: %v", err)
		ui.Render(logParagraph)
	}

	// 移动到文件末尾
	lastOffset, _ := file.Seek(0, io.SeekEnd)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:

				// 移动到上次读取的偏移位置
				file.Seek(lastOffset, io.SeekStart)

				// 读取新内容
				buf := make([]byte, 1024)
				n, _ := file.Read(buf)

				// 更新偏移量
				lastOffset, _ = file.Seek(0, io.SeekCurrent)

				if n > 0 {
					// 更新日志内容
					logParagraph.Text += string(buf[:n])
					if len(strings.Split(logParagraph.Text, "\n")) > 20 { // 限制显示行数
						lines := strings.Split(logParagraph.Text, "\n")
						logParagraph.Text = strings.Join(lines[len(lines)-20:], "\n")
					}
					ui.Render(logParagraph)
				}
			}
		}

	}(ctx)

	// 等待用户退出
	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent && e.ID == "q" {
			cancel()
			break
		}
	}
}

//func viewLogs() {
//	// 初始化终端 UI
//	if err := ui.Init(); err != nil {
//		fmt.Printf("failed to initialize termui: %v\n", err)
//		return
//	}
//	defer ui.Close()
//
//	// 日志文件路径
//	logFile := "example.log"
//
//	// 创建日志显示窗口
//	logParagraph := widgets.NewParagraph()
//	logParagraph.Title = "Log Viewer - Quit with Q"
//	logParagraph.Text = ""
//	logParagraph.SetRect(0, 0, 80, 20)
//	ui.Render(logParagraph)
//
//	// 打开日志文件
//	file, err := os.Open(logFile)
//	if err != nil {
//		logParagraph.Text = fmt.Sprintf("无法打开日志文件: %v", err)
//		ui.Render(logParagraph)
//		return
//	}
//	defer file.Close()
//
//	// 移动到文件末尾
//	file.Seek(0, io.SeekEnd)
//
//	// 上次读取的偏移量
//	var lastOffset int64
//
//	// 创建上下文用于控制协程退出
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	go func(ctx context.Context) {
//		ticker := time.NewTicker(1 * time.Second)
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-ticker.C:
//				// 读取新内容
//				file.Seek(lastOffset, io.SeekStart) // 从上次读取的偏移量开始
//				scanner := bufio.NewScanner(file)
//				var newLines []string
//				for scanner.Scan() {
//					newLines = append(newLines, scanner.Text())
//				}
//
//				// 更新偏移量
//				newOffset, _ := file.Seek(0, io.SeekCurrent)
//				lastOffset = newOffset
//
//				// 显示最后 20 行日志
//				if len(newLines) > 0 {
//					logParagraph.Text += "\n" + fmt.Sprint(newLines)
//					lines := splitToLines(logParagraph.Text)
//					if len(lines) > 20 {
//						logParagraph.Text = joinLastLines(lines, 20)
//					}
//					ui.Render(logParagraph)
//				}
//			}
//		}
//	}(ctx)
//
//	// 等待退出事件
//	for e := range ui.PollEvents() {
//		if e.Type == ui.KeyboardEvent && e.ID == "q" {
//			cancel() // 取消日志读取协程
//			break
//		}
//	}
//}
//
//// 辅助函数：将文本分割为行
//func splitToLines(text string) []string {
//	return strings.Split(text, "\n")
//}
//
//// 辅助函数：取最后 n 行
//func joinLastLines(lines []string, n int) string {
//	if len(lines) <= n {
//		return strings.Join(lines, "\n")
//	}
//	return strings.Join(lines[len(lines)-n:], "\n")
//}
