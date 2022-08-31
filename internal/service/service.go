package service

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/kingcobra2468/cot/internal/config"
)

// Command input arg type.
type ArgType int8

// Command input arg datatype.
type ArgDataType int8

// Command output response type.
type ResponseType int8

// Mapping between the positional index of an argument of the input command and
// its corresponding metadata.
type ArgBindings map[int]*Arg

// The set of supported HTTP methods when sending commands to a client service.
type MethodSet map[string]struct{}

// Mapping between a command input arg type and the corresponding arguments for that
// given arg type.
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

// Commands represents the global schematics of all commands for a given client service.
type Commands struct {
	// All command patterns for a given client service.
	Patterns []string
	// Mapping between a given pattern and command metadata.
	Meta map[string]*Command
}

// Command represents the metadata of a given command. This provides information on how
// and where to perform the HTTP call, how to process the input command, as well
// as how to process the client service output.
type Command struct {
	Method   string
	Endpoint string
	Response Response
	Args     *ArgGroups
}

// Arg represents the metadata about a given input command argument.
type Arg struct {
	TypeInfo
	Type ArgType
	// Whether to compress the rest of the commands from the input command into an array
	// under this argument.
	Compress      bool
	Filter        []interface{}
	FilterEnabled bool
}

// TypeInfo describes the metadata about a given argument and response attribute.
type TypeInfo struct {
	DataType ArgDataType
	Path     string
}

// Response describes how to process the output of the client service and address
// cases of successful and erroneous output.
type Response struct {
	Type    ResponseType
	Success TypeInfo
	Error   TypeInfo
}

// ArgType lists the different ways that the arguments of an input command can
// be parsed.
const (
	// InvalidArg describes an argument whose type that was specified in the configuration
	// but does not exist within cot.
	InvalidArg ArgType = iota
	// QueryArg describes an argument that is designated for the query params.
	QueryArg
	// JsonArg describes an argument that is designated for the JSON body.
	JsonArg
	// EndpointArg describes an argument that will be used to construct the endpoint URL.
	EndpointArg
)

// ResponseType lists the different ways in which the client service command output can
// be processed.
const (
	// InvalidResponse describes a response type that was specified for a given command
	// but does not exist within cot.
	InvalidResponse ResponseType = iota
	// PlainTextResponse describes a response type where the raw response is sent back to
	// the user.
	PlainTextResponse
	// JsonResponse describes a response type where the output will be further processed
	// through JSON.
	JsonResponse
)

// ArgDataType lists the different types of data types that are supported when casting
// the raw input/output into the select types.
const (
	// InvalidType describes a type that was specified but does not exist within cot.
	InvalidType ArgDataType = iota
	// StringType describes that the cast will be made to the string type.
	StringType
	// IntType describes that the cast will be made to the int type.
	IntType
	// FloatType describes that the cast will be made to the float type.
	FloatType
	// BoolType describes that the cast will be made to the bool type.
	BoolType
)

// supportedMethods describes the different HTTP methods that are supported by cot.
var supportedMethods = MethodSet{
	"get":    struct{}{},
	"post":   struct{}{},
	"put":    struct{}{},
	"patch":  struct{}{},
	"delete": struct{}{},
}

// GenerateServices creates a list of services that were specified
// from the configuration file.
func GenerateServices(c *config.Services) ([]Service, error) {
	services := []Service{}
	for _, s := range c.Services {
		service := Service{Name: s.Name, BaseURI: s.BaseURI}
		subCommands := Commands{}
		subCommands.Meta = make(map[string]*Command)
		subCommands.Patterns = []string{}

		for _, cmd := range s.Commands {
			if cmd.Pattern == "" {
				cmd.Pattern = ".*"
			}

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

// generateSubCommand parses and validates the command from the configuration file.
func generateSubCommand(cmdInfo *config.Command) (*Command, error) {
	args, err := generateArgs(cmdInfo.Args, cmdInfo.Method)
	if err != nil {
		return nil, err
	}

	if !methodExists(cmdInfo.Method) {
		return nil, fmt.Errorf("found an invalid method %s", cmdInfo.Method)
	}

	// if no command pattern is specified, then match any pattern
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
	sc.Response.Error = TypeInfo{DataType: edt, Path: cmdInfo.Response.Error.Path}

	return &sc, nil
}

// generateArgs parses and validates the arguments that were specified in the configuration
// file of a given command. The arguments are then preprocessed and aggregated into similar types.
func generateArgs(argInfo *[]config.Arg, method string) (*ArgGroups, error) {
	ag := make(ArgGroups)
	ag[QueryArg] = make(ArgBindings)
	ag[JsonArg] = make(ArgBindings)
	ag[EndpointArg] = make(ArgBindings)
	argCompress := false

	for _, arg := range *argInfo {
		filterEnabled := false
		var filter []interface{}
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

		if argCompress && arg.CompressRest {
			return nil, errors.New("it is not possible to perform arg compression more than once on a single command")
		} else if arg.CompressRest {
			argCompress = true
		}

		if arg.Filter != nil {
			filter = arg.Filter
			filterEnabled = true
		}

		// adds a given argument to a given arg group and points it to the positional
		// index of the input command
		ag[t][arg.Index] = &Arg{Type: t, Compress: arg.CompressRest, TypeInfo: TypeInfo{DataType: dt, Path: arg.Path}, Filter: filter, FilterEnabled: filterEnabled}
	}

	return &ag, nil
}

// parseArgType processes the raw arg type from the configuration file into one of the
// supported types.
func parseArgType(t string) (ArgType, error) {
	switch t {
	case "query":
		return QueryArg, nil
	case "json":
		return JsonArg, nil
	case "endpoint":
		return EndpointArg, nil
	default:
		return InvalidArg, fmt.Errorf("invalid arg type detected \"%s\"", t)
	}
}

// parseArgDataType processes the raw data type into one of the supported data types.
func parseArgDataType(t string) (ArgDataType, error) {
	switch t {
	case "str", "string":
		return StringType, nil
	case "int", "integer":
		return IntType, nil
	case "float", "double":
		return FloatType, nil
	case "bool", "boolean":
		return BoolType, nil
	default:
		return InvalidType, fmt.Errorf("invalid arg datatype detected \"%s\"", t)
	}
}

// parseResponseType processes the raw response type into one of the supported response
// types.
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

// methodExists checks to see if the specified HTTP method is supported by cot.
func methodExists(method string) bool {
	method = strings.ToLower(method)
	_, isValid := supportedMethods[method]

	return isValid
}

// findSubCmd maps the input command into a client service command by doing
// a check of the command pattern.
func (sc Commands) findSubCmd(c *UserInput) (*Command, error) {
	for _, p := range sc.Patterns {
		if found, err := regexp.MatchString(p, c.Raw); err == nil && found {
			return sc.Meta[p], nil
		}
	}

	return nil, errors.New("unable to find a valid subcommand from the input command")
}

func (a Arg) check(ra string) error {
	if !a.FilterEnabled {
		return nil
	}
	for _, val := range a.Filter {
		if val == ra {
			return nil
		}
	}

	return fmt.Errorf("invalid or blacklisted value \"%s\" for arg", ra)
}

// queryString aggregates all of the query arguments from the input command.
func (sc Command) queryString(c *UserInput) (string, error) {
	query := url.Values{}
	argCount := len((*sc.Args)[QueryArg])
	if argCount > len(c.Args) {
		return "", errors.New("unable to parse input command due to invalid amount of input args")
	}

	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[QueryArg] {
		if err := arg.check(c.Args[idx]); err != nil {
			return "", err
		}
		query.Add(arg.Path, c.Args[idx])
		if arg.Compress {
			for i := idx + 1; i < len(c.Args)-1; i++ {
				if err := arg.check(c.Args[i]); err != nil {
					return "", err
				}
				query.Add(arg.Path, c.Args[i])
			}
		}
	}

	return query.Encode(), nil
}

// jsonString aggregates all of the json body arguments from the input command.
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
		var val interface{}
		var err error
		if err := arg.check(c.Args[idx]); err != nil {
			return "", err
		}

		if arg.Compress {
			for i := idx; i < len(c.Args); i++ {
				if err := arg.check(c.Args[i]); err != nil {
					return "", err
				}
				json.ArrayAppendP(c.Args[i], arg.Path)
			}
			break
		}

		switch arg.DataType {
		case StringType:
			val = c.Args[idx]
		case IntType:
			val, err = strconv.ParseInt(c.Args[idx], 10, 64)
			if err != nil {
				return "", fmt.Errorf("unable to parse arg %s into an int", c.Args[idx])
			}
		case FloatType:
			val, err = strconv.ParseFloat(c.Args[idx], 64)
			if err != nil {
				return "", fmt.Errorf("unable to parse arg %s into an float", c.Args[idx])
			}
		case BoolType:
			val, err = strconv.ParseBool(c.Args[idx])
			if err != nil {
				return "", fmt.Errorf("unable to parse arg %s into an boolean", c.Args[idx])
			}
		}

		json.SetP(val, arg.Path)
	}

	return json.String(), nil
}

// endpointString aggregates all of the endpoint arguments into the endpoint URL
func (sc Command) endpointString(c *UserInput) (string, error) {
	endpoint := []string{}
	argCount := len((*sc.Args)[EndpointArg])
	if argCount > len(c.Args) {
		return "", errors.New("unable to parse input command due to invalid among of input args")
	}

	if argCount == 0 {
		return "", nil
	}

	for idx, arg := range (*sc.Args)[EndpointArg] {
		if err := arg.check(c.Args[idx]); err != nil {
			return "", err
		}
		endpoint = append(endpoint, c.Args[idx])
	}

	return strings.Join(endpoint, "/"), nil
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

// setupRequest prepares for the client service command request by parsing the user command.
// Preprocessing is then preformed to prepare the request based on the criteria specified for
// the command.
func (s Service) setupRequest(c *Command, ui *UserInput) (*http.Request, error) {
	query, err := c.queryString(ui)
	if err != nil {
		return nil, err
	}
	json, err := c.jsonString(ui)
	if err != nil {
		return nil, err
	}
	endpoint, err := c.endpointString(ui)
	if err != nil {
		return nil, err
	}

	var serializedJson *bytes.Buffer = bytes.NewBuffer([]byte{})
	// check if any json args exist for the given command
	if len(json) > 0 {
		serializedJson = bytes.NewBuffer([]byte(json))
	}

	req, err := http.NewRequest(strings.ToUpper(c.Method), s.BaseURI+path.Join(c.Endpoint, endpoint), serializedJson)
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

// processResponse processes the client service command output based on the criteria
// specified for the command.
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
	case http.StatusOK:
		respPath = c.Response.Success.Path
	default:
		respPath = c.Response.Error.Path
	}
	msg := output.Path(respPath).String()

	return msg, nil
}
