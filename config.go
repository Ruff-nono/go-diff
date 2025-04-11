package main

import (
	"embed"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var defaultAgentFS embed.FS

var config *Config

// EnvVarRegex 环境变量占位符正则表达式，匹配 ${ENV:default} 格式
var EnvVarRegex = regexp.MustCompile(`\${(?P<ENV>[A-Z0-9_]+):?(?P<DEF>.*)}`)

type Config struct {
	SelfPort              StringValue `yaml:"self_port"`
	Host1                 StringValue `yaml:"host1"`
	Host2                 StringValue `yaml:"host2"`
	PathPattern           StringValue `yaml:"path_pattern"`
	ErrorMaxQueue         StringValue `yaml:"path_pattern"`
	HeadersInclude        StringValue `yaml:"headers_include"`
	CompareStatusCode     StringValue `yaml:"compare_status_code"`
	EquivalentStatusCodes StringValue `yaml:"equivalent_status_codes"`
	CompareBody           StringValue `yaml:"compare_body"`
	BodiesExclude         StringValue `yaml:"bodies_exclude"`
}

func init() {
	defaultConfig, err := defaultAgentFS.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to open config file: %v\n", err)
		return
	}
	if err1 := yaml.Unmarshal(defaultConfig, &config); err1 != nil {
		log.Fatalf("Failed to unmarshal config file: %v\n", err)
		return
	}

	return
}

// StringValue 配置值封装（支持环境变量解析）
type StringValue struct {
	rawValue string
	envKey   string
	defValue string
}

// UnmarshalYAML 实现自定义解析逻辑
func (s *StringValue) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}
	s.rawValue = str

	// 解析环境变量占位符
	if matches := EnvVarRegex.FindStringSubmatch(str); len(matches) >= 3 {
		s.envKey = matches[1]
		s.defValue = matches[2]
	} else {
		s.defValue = str
	}
	return nil
}

func (s *StringValue) String() string {
	if v := os.Getenv(s.envKey); v != "" {
		return v
	}
	return s.defValue
}

func (s *StringValue) Bool() bool {
	v, _ := strconv.ParseBool(s.String())
	return v
}

func (s *StringValue) Int() int {
	v, _ := strconv.Atoi(s.String())
	return v
}

func (s *StringValue) StringSlice() []string {
	return splitString(s.String(), ",")
}

func (s *StringValue) IntSliceSlice() [][]int {
	var result [][]int
	for _, group := range splitString(s.String(), ";") {
		var nums []int
		for _, item := range splitString(group, ",") {
			if n, err := strconv.Atoi(item); err == nil {
				nums = append(nums, n)
			}
		}
		if len(nums) > 0 {
			result = append(result, nums)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
