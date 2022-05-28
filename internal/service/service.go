package service

import (
	"bytes"
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
type ArgDataType int8
type ResponseType int8
type ArgBindings map[int]*Arg
type MethodSet map[string]struct{}
type ArgGroups map[ArgType]ArgBindings

// Service handles the communication between a command request and the associated
// client service.
type Service struct {
	Commands
	Name    string
	BaseURI string
}

// UserInput sets up the schema for a command request via the name of a command and its
// its arguments.
type UserInput struct {
	Name string
	Args []string
	Raw  string
}

type Commands struct {
	Patterns []string
	Meta     map[string]*Command
}

type Command struct {
	Method   string
	Endpoint string
	Response Response
	Args     *ArgGroups
}

type Arg struct {
	TypeInfo
	Type     ArgType
	Compress bool
}

type TypeInfo struct {
	DataType ArgDataType
	Path     string
}

type Response struct {
	Type    ResponseType
	Success TypeInfo
	Error   TypeInfo
}

const (
	InvalidArg ArgType = iota
	QueryArg
	JsonArg
)

const (
	InvalidResponse ResponseType = iota
	PlainTextResponse
	JsonResponse
)

const (
	InvalidType ArgDataType = iota
	StringType
	IntType
	FloatType
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
		service := Service{Name: s.Name, BaseURI: s.BaseURI}
		subCommands := Commands{}
		subCommands.Meta = make(map[string]*Command)
		subCommands.Patterns = []string{}

		for _, cmd := range s.Commands {
			subCommands.Patterns = append(subCommands.Patterns, cmd.Pattern)
			sc, err := generateSubCommand(cmd)
			if err != nil {
				return nil, err
			}

			if _, exists := subCommands.Meta[cmd.Pattern]; exists {
				return nil, fmt.Errorf("repeated pattern detected in service \"%s\"", s.Name)
			}
			subCommands.Meta[cmd.Pattern] = sc
		}

		service.Commands = subCommands
		services = append(services, service)
	}

	return services, nil
}

func generateSubCommand(cmdInfo *config.Command) (*Command, error) {
	args, err := generateArgs(cmdInfo.Args, cmdInfo.Method)
	if err != nil {
		return nil, err
	}

	if !methodExists(cmdInfo.Method) {
		return nil, fmt.Errorf("found an invalid method %s", cmdInfo.Method)
	}

	if len(cmdInfo.Pattern) == 0 {
		cmdInfo.Pattern = ".*"
	}

	sc := Command{Endpoint: cmdInfo.Endpoint, Method: cmdInfo.Method, Args: args}
	rt, err := parseResponseType(cmdInfo.Response.Type)
	if err != nil {
		return nil, err
	}

	sc.Response.Type = rt
	if rt == PlainTextResponse {
		return &sc, nil
	}

	sdt, err := parseArgDataType(cmdInfo.Response.Success.DataType)
	if err != nil {
		return nil, err
	}
	sc.Response.Success = TypeInfo{DataType: sdt, Path: cmdInfo.Response.Success.Path}

	edt, err := parseArgDataType(cmdInfo.Response.Error.DataType)
	if err != nil {
		return nil, err
	}
	sc.Response.Success = TypeInfo{DataType: edt, Path: cmdInfo.Response.Error.Path}

	return &sc, nil
}

func generateArgs(argInfo *[]config.Arg, method string) (*ArgGroups, error) {
	fmt.Println(argInfo)
	ag := make(ArgGroups)
	ag[QueryArg] = make(ArgBindings)
	ag[JsonArg] = make(ArgBindings)

	for _, arg := range *argInfo {
		t, err := parseArgType(arg.Type)
		if err != nil {
			return nil, err
		}
		dt, err := parseArgDataType(arg.DataType)
		if err != nil {
			return nil, err
		}
		if t == JsonArg && strings.EqualFold("get", method) {
			return nil, fmt.Errorf("arg index %d for path %s cannot exist for GET requests", arg.Index, arg.Path)
		}

		ag[t][arg.Index] = &Arg{Type: t, Compress: arg.CompressRest, TypeInfo: TypeInfo{DataType: dt, Path: arg.Path}}
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
		return InvalidArg, fmt.Errorf("invalid arg type detected \"%s\"", t)
	}
}

func parseArgDataType(t string) (ArgDataType, error) {
	switch t {
	case "str", "string":
		return StringType, nil
	case "int", "integer":
		return IntType, nil
	case "float", "double":
		return FloatType, nil
	default:
		return InvalidType, fmt.Errorf("invalid arg datatype detected \"%s\"", t)
	}
}

func parseResponseType(t string) (ResponseType, error) {
	switch t {
	case "plain_text":
		return PlainTextResponse, nil
	case "json":
		return JsonResponse, nil
	default:
		return InvalidResponse, fmt.Errorf("invalid response type detected \"%s\"", t)
	}
}

func methodExists(method string) bool {
	method = strings.ToLower(method)
	_, isValid := supportedMethods[method]

	return isValid
}

func (sc Commands) findSubCmd(c *UserInput) (*Command, error) {
	for _, p := range sc.Patterns {
		if found, err := regexp.MatchString(p, c.Raw); err == nil && found {
			return sc.Meta[p], nil
		}
	}

	return nil, errors.New("unable to find a valid subcommand from the input command")
}

func (sc Command) queryString(c *UserInput) (string, error) {
	query := url.Values{}
	argCount := len((*sc.Args)[QueryArg])
	if argCount > len(c.Args) {
		return "", errors.New("unable to parse input command due to invalid among of input args")
	}

	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[QueryArg] {
		query.Add(arg.Path, c.Args[idx])
	}

	return query.Encode(), nil
}

func (sc Command) jsonString(c *UserInput) (string, error) {
	json := gabs.New()
	argCount := len((*sc.Args)[JsonArg])
	if argCount > len(c.Args) {
		return "", errors.New("unable to parse input command due to invalid among of input args")
	}
	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[JsonArg] {
		json.SetP(c.Args[idx], arg.Path)
	}

	return json.String(), nil
}

// Execute will push the command request to the associated client service and will
// retrieve the output.
func (s Service) Execute(ui *UserInput) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}
	c, err := s.findSubCmd(ui)
	if err != nil {
		return "", err
	}

	req, err := s.setupRequest(c, ui)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	msg, err := s.processResponse(c, resp)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (s Service) setupRequest(c *Command, ui *UserInput) (*http.Request, error) {
	query, err := c.queryString(ui)
	if err != nil {
		return nil, err
	}
	json, err := c.jsonString(ui)
	if err != nil {
		return nil, err
	}

	var serializedJson *bytes.Buffer
	if len(json) != 0 {
		serializedJson = bytes.NewBuffer([]byte(json))
	}
	fmt.Println("Q ", query)
	fmt.Println("J ", json)
	req, err := http.NewRequest(c.Method, s.BaseURI+c.Endpoint, serializedJson)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query
	if len(json) != 0 {
		req.Header.Add("Content-Type", "application/json")
	}

	switch c.Response.Type {
	case PlainTextResponse:
		req.Header.Add("Accept", "text/plain")
	case JsonResponse:
		req.Header.Add("Accept", "application/json")
	}

	return req, nil
}

func (s Service) processResponse(c *Command, resp *http.Response) (string, error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if c.Response.Type == PlainTextResponse {
		return string(bodyBytes), nil
	}

	output, err := gabs.ParseJSON(bodyBytes)
	if err != nil {
		return "", err
	}

	var respPath string
	switch resp.StatusCode {
	case 200:
		respPath = c.Response.Success.Path
	default:
		respPath = c.Response.Error.Path
	}

	msg, ok := output.Path(respPath).Data().(string)
	if !ok {
		return "", errors.New("unable to parse output json")
	}

	return msg, nil
}
