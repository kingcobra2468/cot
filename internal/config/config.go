// config handles the parsing of cot config into useful
// structures. Also handles the setup of components.
package config

// Services contains configuration on each of the services and the client
// numbers authorized to use it. Also contains the GVoice number
// bindings that the client numbers send messages to. The ability to
// see whether text encryption is enabled is also accessible.
type Services struct {
	Services       []*Service `mapstructure:"services"`
	GVoiceNumber   string     `mapstructure:"gvoice_number"`
	TextEncryption bool       `mapstructure:"text_encryption"`
}

// Service contains configuration on the service name (also used as the command name)
// as well as the service base URI. Also contains a list of client numbers authorized
// to use this service.
type Service struct {
	Name          string     `mapstructure:"name"`
	BaseURI       string     `mapstructure:"base_uri"`
	ClientNumbers []string   `mapstructure:"client_numbers"`
	Commands      []*Command `mapstructure:"commands"`
}

// Command contains the signature for each of the subcommands. This includes the pattern
// that determines if the command exists from the user input, as well as various metadata
// in regards to how to send that command to a given client service.
type Command struct {
	Pattern  string   `mapstructure:"pattern"`
	Method   string   `mapstructure:"method"`
	Endpoint string   `mapstructure:"endpoint"`
	Args     *[]Arg   `mapstructure:"args"`
	Response Response `mapstructure:"response"`
}

// Arg represents argument config for a given command of a given client service.
type Arg struct {
	TypeInfo     `mapstructure:",squash"`
	Index        int    `mapstructure:"index"`
	Type         string `mapstructure:"type"`
	CompressRest bool   `mapstructure:"compress_rest"`
}

// Response contains the configuration of the response signature of a given command.
type Response struct {
	Type    string   `mapstructure:"type"`
	Success TypeInfo `mapstructure:"success"`
	Error   TypeInfo `mapstructure:"error"`
}

// TypeInfo represents type info metadata for a given argument or response type.
type TypeInfo struct {
	Path     string `mapstructure:"path"`
	DataType string `mapstructure:"datatype"`
}

// Encryption contains configuration on various options for encryption and files
// needed if PGP encryption is enabled.
type Encryption struct {
	TextEncryption           bool   `mapstructure:"text_encryption"`
	SignatureVerification    bool   `mapstructure:"sig_verification"`
	Base64Encoding           bool   `mapstructure:"base64_encoding"`
	PublicKeyFile            string `mapstructure:"public_key_file"`
	PrivateKeyFile           string `mapstructure:"private_key_file"`
	Passphrase               string `mapstructure:"passphrase"`
	ClientNumberPublicKeyDir string `mapstructure:"cn_public_key_dir"`
}

// GVMS contains configuration on how to communicate with GVMS server.
type GVMS struct {
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
}
