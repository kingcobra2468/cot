// config handles the parsing of cot config into useful
// structures. Also handles the setup of components.
package config

// Services contains configuration on each of the services and the client
// numbers authorized to use it. Also contains the GVoice number
// bindings that the client numbers send messages to. The ability to
// see whether text encryption is enabled is also accessible.
type Services struct {
	Services       []Service `mapstructure:"services"`
	GVoiceNumber   string    `mapstructure:"gvoice_number"`
	TextEncryption bool      `mapstructure:"text_encryption"`
}

// Service contains configuration on the service name (also used as the command name)
// as well as the service base URI. Also contains a list of client numbers authorized
// to use this service.
type Service struct {
	Name          string   `mapstructure:"name"`
	BaseURI       string   `mapstructure:"base_uri"`
	ClientNumbers []string `mapstructure:"client_numbers"`
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
