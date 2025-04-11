package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/event"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"regexp"
)

// Prometheus 监控指标
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"api", "state"},
	)

	pathPatterns []PathPattern
)

func init() {
	// 注册 Prometheus 指标
	prometheus.MustRegister(requestsTotal)
	pathPatterns = InitPathPatterns()
}

type PathPattern struct {
	raw   string         // 原始配置如 /api/user/:userId
	regex *regexp.Regexp // 转换后的正则表达式
}

func InitPathPatterns() []PathPattern {
	regexPatterns := config.PathPattern.StringSlice()
	compiled := make([]PathPattern, 0, len(regexPatterns))
	for _, p := range regexPatterns {
		compiled = append(compiled, PathPattern{
			raw:   p,
			regex: regexp.MustCompile(p),
		})
	}
	return compiled
}

func matchPath(api string) string {
	for _, pattern := range pathPatterns {
		if pattern.regex.MatchString(api) {
			return pattern.raw
		}
	}
	return api
}

func recordRequest(r *http.Request, errorType string, details string) {
	api := matchPath(r.URL.Path)
	logger.Printf("recordRequest: api:[%s], errorType:%s,\n details:%s", api, errorType, details)
	requestsTotal.WithLabelValues(api, errorType).Inc()
	recordCurlLog(r, api, errorType, details)
}

// combinedChartHandler 收集指标数据，根据每个 API 生成一个 BarChart，并按竖向合并所有图表
func combinedChartHandler(w http.ResponseWriter, _ *http.Request) {
	metrics := getRealMetrics()
	apiData := groupByAPI(metrics)

	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)

	// 3. 为每个API创建饼图
	for api, data := range apiData {
		handler := fmt.Sprintf(`
        (params) => {
            const apiName = "%s";
 			const currentHost = window.location.protocol + "//" + window.location.host;
			window.open(currentHost + '/debug/errors?api=' + encodeURIComponent(apiName) + 
                '&state=' + encodeURIComponent(params.name)
            );
        }
    `, api)
		pie := charts.NewPie()
		pie.SetGlobalOptions(
			charts.WithTitleOpts(opts.Title{
				Title:    "API: " + api,
				Subtitle: "State Distribution",
			}),
			charts.WithLegendOpts(opts.Legend{
				Left: "right",
			}),
			charts.WithEventListeners(
				event.Listener{
					EventName: "click",
					Handler:   opts.FuncOpts(handler),
				},
			),
		)

		// 添加数据并设置样式
		pie.AddSeries("state", data).
			SetSeriesOptions(
				charts.WithLabelOpts(opts.Label{
					Formatter: "{b}: {c} ({d}%)",
				}),
				charts.WithPieChartOpts(opts.PieChart{
					Radius: []string{"40%", "75%"},
				}),
			)

		page.AddCharts(pie)
	}

	// 3.4 渲染到HTTP响应
	w.Header().Set("Content-Type", "text/html")
	_ = page.Render(w) // 直接写入ResponseWriter
}

type MetricData struct {
	API   string
	State string
	Value float64
}

func getRealMetrics() []MetricData {
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic(err)
	}
	var metrics []MetricData
	// 遍历指标集合
	for _, mf := range metricFamilies {
		// 筛选目标指标 requests_total
		if *mf.Name == "requests_total" {
			for _, metric := range mf.Metric {
				var api, state string
				var value float64

				// 解析标签（api 和 state）
				for _, label := range metric.GetLabel() {
					switch *label.Name {
					case "api":
						api = *label.Value
					case "state":
						state = *label.Value
					}
				}

				// 解析计数器值
				if metric.Counter != nil {
					value = *metric.Counter.Value
				}

				metrics = append(metrics, MetricData{
					API:   api,
					State: state,
					Value: value,
				})
			}
		}
	}
	return metrics
}

func groupByAPI(metrics []MetricData) map[string][]opts.PieData {
	dataMap := make(map[string][]opts.PieData)
	for _, m := range metrics {
		dataMap[m.API] = append(dataMap[m.API], opts.PieData{
			Name:  m.State,
			Value: m.Value,
		})
	}
	return dataMap
}
