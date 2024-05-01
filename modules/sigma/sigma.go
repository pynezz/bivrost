package sigma

// type Rule struct {
// 	// Required fields
// 	Title     string
// 	Logsource Logsource
// 	Detection Detection

// 	ID          string        `yaml:",omitempty" json:",omitempty"`
// 	Related     []RelatedRule `yaml:",omitempty" json:",omitempty"`
// 	Status      string        `yaml:",omitempty" json:",omitempty"`
// 	Description string        `yaml:",omitempty" json:",omitempty"`
// 	Author      string        `yaml:",omitempty" json:",omitempty"`
// 	Level       string        `yaml:",omitempty" json:",omitempty"`
// 	References  []string      `yaml:",omitempty" json:",omitempty"`
// 	Tags        []string      `yaml:",omitempty" json:",omitempty"`

// 	// Any non-standard fields will end up in here
// 	AdditionalFields map[string]interface{} `yaml:",inline,omitempty" json:",inline,omitempty"`
// }

// /*
// To help understand what the above code snippet accomplishes, this Sigma rule will be separated into three main components:

// Detection (required)
// What malicious behaviour the rule searching for.
// The detection section is the most important component of any Sigma rule. It specifies exactly what the rule is looking for across relevant logs.

// Logsource (required)
// What types of logs this detection should search over.
// Metadata (optional)
// Other information about the detection.
// */

// // // // Sigma is a struct that holds the Sigma rules and the Sigma parser
// // // type Sigma struct {
// // // 	Rules  []sigma.Rule
// // // 	Parser sigma.Evaluator
// // // }

// // func init() {
// // 	// You can load/create rules dynamically or use sigmac to load Sigma rule files
// // 	var rule, _ = sigma.ParseRule(contents)

// // 	// Rules need to be wrapped in an evaluator.
// // 	// This is also where (if needed) you provide functions implementing the count, max, etc. aggregation functions
// // 	e := sigma.Evaluator(rule, options...)
// // }

// // // NewSigma creates a new Sigma struct
// // func NewSigma() *Sigma {
// // 	return &Sigma{
// // 		Parser: sigma.RuleEvaluator{
// // 			Options: evaluator.Option,
// // 		},
// // 	}
// // }

// // // LoadRules loads the Sigma rules from the given path
// // func (s *Sigma) LoadRules(path string) error {
// // 	rules, err := sigma.LoadRules(path)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	s.Rules = rules
// // 	return nil
// // }

// // // Parse parses the given log and returns the parsed logs
// // func (s *Sigma) Parse(log string) ([]sigma.ParsedRule, error) {
// // 	return s.Parser.Parse(log, s.Rules)
// // }

// type Rule struct {
// 	RuleEvaluator
// }

// func parseRule(rule string) (Rule, error) {
// 	return Rule{}, nil
// }
