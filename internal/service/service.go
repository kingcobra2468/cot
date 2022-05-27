package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/kingcobra2468/cot/internal/config"
)

type ArgType int8
type ArgMapper map[int]*Arg
type ResponseType int8
type MethodSet map[string]struct{}
type ArgGroups map[ArgType]ArgMapper

// Service handles the communication between a command request and the associated
// client service.
type Service struct {
	SubCommands
	Name              string
	BaseURI           string
	Endpoint          string
	Method            string
	HandleSubCommands bool
}

// Command sets up the schema for a command request via the name of a command and its
// its arguments.
type Command struct {
	Name      string
	Arguments []string
	RawInput  string
}

type SubCommands struct {
	Patterns []string
	Meta     map[string]*SubCommand
}

type SubCommand struct {
	Method   string
	Endpoint string
	Response Response
	Args     *ArgGroups
}

type Arg struct {
	Type      ArgType
	Namespace string
}

type Response struct {
	Type      ResponseType
	Namespace string
}

// CommandRequest contains the JSON request schema.
type CommandRequest struct {
	Args []string `json:"args"`
}

// CommandRequest sets up the JSON response schema.
type CommandResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error,omitempty"`
}

const (
	InvalidArgType ArgType = iota
	QueryArg
	JsonArg
)

const (
	InvalidResponseType ResponseType = iota
	PlainTextResponse
	JsonResponse
)

var supportedMethods = MethodSet{
	"get":    struct{}{},
	"post":   struct{}{},
	"put":    struct{}{},
	"patch":  struct{}{},
	"delete": struct{}{},
}

// GenerateServices creates a list of services that were specified
// in the configuration file.
func GenerateServices(c *config.Services) ([]Service, error) {
	services := []Service{}
	for _, s := range c.Services {
		service := Service{Name: s.Name, BaseURI: s.BaseURI, Endpoint: s.Endpoint}
		if s.Commands == nil {
			services = append(services, service)
			continue
		}

		subCommands := SubCommands{}
		subCommands.Meta = make(map[string]*SubCommand)
		subCommands.Patterns = []string{}
		for _, cmd := range s.Commands {
			if cmd.Args == nil {
				return nil, errors.New("subcommand ambiguity detected as a result of no arg being present")
			}

			subCommands.Patterns = append(subCommands.Patterns, cmd.Pattern)
			sc, err := generateSubCommand(cmd)
			if err != nil {
				return nil, err
			}

			subCommands.Meta[cmd.Pattern] = sc
		}

		service.SubCommands = subCommands
		service.HandleSubCommands = true
		services = append(services, service)
	}

	return services, nil
}

func generateSubCommand(cmdInfo *config.Command) (*SubCommand, error) {
	args, err := generateArgs(&cmdInfo.Args, cmdInfo.Method)
	if err != nil {
		return nil, err
	}

	if !methodExists(cmdInfo.Method) {
		return nil, fmt.Errorf("found an invalid method %s", cmdInfo.Method)
	}

	sc := SubCommand{Endpoint: cmdInfo.Endpoint, Method: cmdInfo.Method, Args: args}
	if len(cmdInfo.Response.Type) == 0 {
		sc.Response.Namespace = "message"
		sc.Response.Type = JsonResponse
	}
	rt, err := parseResponseType(cmdInfo.Response.Type)
	if err != nil {
		return nil, err
	}

	sc.Response.Type = rt
	sc.Response.Namespace = cmdInfo.Response.Namespace

	return &sc, nil
}

func generateArgs(argInfo *map[int]*config.TypeInfo, method string) (*ArgGroups, error) {
	ag := make(ArgGroups)
	ag[QueryArg] = make(ArgMapper)
	ag[JsonArg] = make(ArgMapper)

	for idx, arg := range *argInfo {
		t, err := parseArgType(arg.Type)
		if err != nil {
			return nil, err
		}
		if t == JsonArg && strings.EqualFold("get", method) {
			return nil, fmt.Errorf("arg index %d for namespace %s cannot exist for GET requests", idx, arg.Namespace)
		}

		ag[t][idx] = &Arg{Type: t, Namespace: arg.Namespace}
	}

	return &ag, nil
}

func parseArgType(t string) (ArgType, error) {
	switch t {
	case "query":
		return QueryArg, nil
	case "json":
		return JsonArg, nil
	default:
		return InvalidArgType, fmt.Errorf("invalid arg type detected \"%s\"", t)
	}
}

func parseResponseType(t string) (ResponseType, error) {
	switch t {
	case "plain_text":
		return PlainTextResponse, nil
	case "json":
		return JsonResponse, nil
	default:
		return InvalidResponseType, fmt.Errorf("invalid response type detected \"%s\"", t)
	}
}

func methodExists(method string) bool {
	method = strings.ToLower(method)
	_, isValid := supportedMethods[method]

	return isValid
}

func (sc SubCommands) findSubCmd(c *Command) (*SubCommand, error) {
	for _, p := range sc.Patterns {
		if found, err := regexp.MatchString(p, c.RawInput); err == nil && found {
			return sc.Meta[p], nil
		}
	}

	return nil, errors.New("unable to find a valid subcommand from the input command")
}

func (sc SubCommand) queryString(c *Command) (string, error) {
	query := url.Values{}
	argCount := len((*sc.Args)[QueryArg])
	if argCount > len(c.Arguments) {
		return "", errors.New("unable to parse input command due to invalid among of input args")
	}

	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[QueryArg] {
		query.Add(arg.Namespace, c.Arguments[idx])
	}

	return query.Encode(), nil
}

func (sc SubCommand) jsonString(c *Command) (string, error) {
	json := gabs.New()
	argCount := len((*sc.Args)[JsonArg])
	if argCount > len(c.Arguments) {
		return "", errors.New("unable to parse input command due to invalid among of input args")
	}
	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[JsonArg] {
		json.SetP(c.Arguments[idx], arg.Namespace)
	}

	return json.String(), nil
}

// Execute will push the command request to the associated client service and will
// retrieve the output.
func (s Service) Execute(c *Command) (string, error) {

}

func (s Service) executeCommand(c *Command) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}

	data, err := json.Marshal(CommandRequest{Args: c.Arguments})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.BaseURI+s.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var output CommandResponse
	err = json.Unmarshal(bodyBytes, &output)
	// check if output parsable
	if err != nil {
		return "", err
	}
	// check if error was sent back from client service
	if output.Error != nil {
		return "", output.Error
	}

	return output.Message, err
}

func (s Service) executeSubCommand(c *Command) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}
	sc, err := s.findSubCmd(c)
	if err != nil {
		return "", err
	}
	query, err := sc.queryString(c)
	if err != nil {
		return "", err
	}
	json, err := sc.jsonString(c)
	if err != nil {
		return "", err
	}

	var serializedJson *bytes.Buffer
	if len(json) != 0 {
		serializedJson = bytes.NewBuffer([]byte(json))
	}

	req, err := http.NewRequest(sc.Method, s.BaseURI+s.Endpoint+sc.Endpoint, serializedJson)
	req.URL.RawQuery = query
	req.Header.Add("Content-Type", "application/json")

	switch sc.Response.Type {
	case PlainTextResponse:
		req.Header.Add("Accept", "text/plain")
	case JsonResponse:
		req.Header.Add("Accept", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if sc.Response.Type == PlainTextResponse {
		return string(bodyBytes), nil
	}

	output, err := gabs.ParseJSON(bodyBytes)
	if err != nil {
		return "", err
	}

	msg, ok := output.Path(sc.Response.Namespace).Data().(string)
	if !ok {
		return "", errors.New("unable to parse output json")
	}
	return msg, nil
}
