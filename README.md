# **COT**
Commands over text(COT) is platform for bridging internal services and mobile users
over the SMS/MMS protocol. Through a rich and programmable interface, one is able to define
and fine-tune commands designated to internal services without doing any port-forwarding. 

COT's main benefit arises with the ability to interact with internal services without
the need to do any port forwarding or exposure to the public internet. Thus, with COT
running on the same network as the internal services, commands can be created along with
client service wrappers which will accept requests from cot and then further pass them
on to the internal services.

## **Terminology**
- **Client Number=** phone number that sends the command request.
- **GVoice Number=** Gvoice phone number that services as receiver for commands that
  come from client numbers.
- **GVMS=** microservice for interacting with GVoice's APIs.
- **Client Service=** user-defined service that is implemented externally of COT and is
  responsible for taking in a command request, executing it, and returning the response.

## **Architecture**
![photo](images/cot.jpg)

COT, being generic, enables one to create any command they want as long as one defines
it within `cot_sm.yaml`. Within the config file, one would specify:
1. The name of the service and a list of commands that hit various endpoints of such service.
2. List of client numbers authorized to use the command.

## **Flow**
The end-to-end flow of COT is as follows:
1. COT initializes a worker that checks the (GVoice Number, Client Number)
   link periodically by polling GVMS. *By listening to only the subnet of defined client numbers,
   COT, by nature, will ignore all numbers that have not been whitelisted by any of the
   services within `cot_sm.yaml`.*
2. A client sends a command in the format "[cmd] [arg 0] [arg 1] ... [arg N]" (split into tokens by 
   the space delimiter) to a GVoice number that COT polls.
3. On detection of a new user command, COT parsers the command and checks if the client number is
   authorized to run this command. Non-authorized commands will be rejected. Likewise, COT also checks
   if the command exists.
4. COT tries to match the user command to a given command of a service by doing pattern checking against
   the input. In the case were no patterns match, an error is returned. Otherwise, the arguments are
   reformated into appropriate arg groups and the command is sent to the configured service + endpoint along
   with the defined HTTP method. 
5. The output of the command would then parsed based on the response configuration and is then sent back
   to the client number.

## **Encryption**
![photo](images/cot_encryption.jpg)

Given that the SMS/MMS protocol is the foundation for COT, all messages will be visible
by default. This includes but is not limited to ATT, Verizon, Google (due to GVoice), among
other parties. Thus, this pushes the need for encryption.

As seen in the diagram, COT features 3 main flows, though further tweaking is possible with some
limitations.

### **Flows**

#### **No Encryption Flow**
This is the least secure of all flows and should be used in the case where the 2 other flows are
not viable.

#### **PGP Encryption Flow**
This flow requires the client number to sign the command request with COT's public key prior to
sending the message. The ASCII armored message will then be sent to COT. COT will send the response
encrypted with the client number's public key.

#### **PGP Encryption Flow & Base64**
This flow requires the client number to sign the command request with COT's public key prior to
sending the message. Afterwards, the ASCII armored message would then need to be base64 encoded.
In the response message, COT will do the same and base64 encode the ASCII armoured message. The 
reason behind base64 encoding is due to some MMS/SMS clients on Android doing compression. Even
though the compression might seem harmless, PGP requires certain schema and this will render the
message useless. Thus, using base64 will preserve all the formatting in order to prevent corruption.

### **Note on PGP Signatures**
For the *PGP Encryption* and *PGP Encryption & Base64* flows, the option to set
`COT_SIG_VERIFICATION` is possible. This will validate that the input command is signed by
the client number and will also sign the response with COT's private key. However, some SMS/MMS
phone clients contain message size caps (even though MMS theoretically supports 5Mb messages), and
thus this option should might not work everywhere. Thus, unless tested that it works for your needs,
`COT_SIG_VERIFICATION` should remain as `false`. If set to false, avoid ensure that signing is off
from the PGP client app used on the client number phone, in order to avoid large message sizes.

### **Note on PGP and Base64 Encoding From Client Side**
Since there is no user client for COT as it is intended to use the default SMS/MMS client, one
would have to download a PGP encoding/decoding as well as a base64 encoding/decoding app from the
App Store/Play Store and do the steps themselves each time.

#### **PGP Encryption Flow**
1. Encrypt command with PGP app.
2. Send output to COT via SMS/MMS client.
3. After response arrives from COT, paste it into PGP app's decoder and see command output.

#### **PGP Encryption & Base64 Flow**
1. Encrypt command with PGP app.
2. Paste encrypted message into encoder of Base64 app.
3. Send output to COT via SMS/MMS client.
4. After response arrives from COT, paste it into decoder of Base64 app.
5. Copy output and paste into PGP app's decoder and see command output.

## **Configuration**

### **General Configuration**
Most configuration is done via the `cot_sm.yaml` file which needs to be copied/renamed from 
`cot_sm_template.yaml`. By default, the file needs to be located in the same directory as
the executable or `main.go`, unless an alternate path is specified via the `COT_CONF_DIR`
environment variable.

#### **Sample Configuration**
The best way to explain the configuration achievable with COT is through an example `cot_sm.yaml`
file. Note that none of the data is real and the numbers are fake and should not be contacted. 
```yaml
---
gvms:
  hostname: "192.168.1.10"
  port: 7777
gvoice_number: 11111111111
services:
  - name: car
    base_uri: "http://192.168.1.11:8086"
    client_numbers:
      - 12222222222
    commands:
      - args:
          - datatype: str
            type: endpoint
            index: 1
          - datatype: int
            path: price.current
            type: json
            index: 2
            compress_rest: true
        endpoint: /cars
        method: put
        pattern: .*changeprice.*
        response:
          type: json
          success:
            datatype: str
            path: data
          error:
            datatype: str
            path: data
      - args:
          - datatype: str
            type: endpoint
            index: 1
        endpoint: /cars
        method: delete
        pattern: .*remove.*
        response:
          type: json
          success:
            datatype: str
            path: data
          error:
            datatype: str
            path: data
```
The configuration above defines a single service. To send a command, a client number must
send a text to `11111111111` which COT polls from GVMS. Here, GVMS runs on `192.168.1.10`
with port `7777`. The single service, is represented by the command `car`. 

The command `car` exposes two subcommands, one of which gets triggered when the user command
contains `changeprice` while the other gets triggered when the user command contains `remove`.

The first subcommand expects the user command to be `car changeprice [carname] price_1 price_2 ... price_n`.
From the mapping, the raw command will be converted into a PUT request with the URL being
`http://192.168.1.11:8086/cars/[carname]` with the payload being
`{"price": {"current": [price_1, price_2, ..., price_n]}}`. The command will then return the contents of `data`
key from the response JSON payload.

The seconds subcommand expects the user command to be `car remove [carname]`. From the mapping, the raw command
will be converted into a DELETE request with the URL being `http://192.168.1.11:8086/cars/`. The command will
then return the contents of `data` key from the response JSON payload.

#### **Reference**
The reference contains all of possible fields within `cot_sm.yaml`. The fields are represented by their path
from the root of the config file. So, for example `gvms.hostname` will represent:
```yaml
gvms:
  hostname:
```
It is also important to know how COT does user command input parsing. COT parses the raw user input string into
tokens that are split by space character. The first token corresponds to the service/base command name. The
rest of the tokens correspond to args and are 0 indexed. When doing pattern patching of the command, the matching
is done against the complete raw input, including the service/base command name.

- **gvms.hostname** The hostname for GVMS.
- **gvms.port** The port for GVMS.
- **gvoice_number** The google voice number that client numbers need to send commands to
  in order to be picked up by COT.
- **services[].base_url** The base url used in the construction of an endpoint for a given service.
- **services[].client_numbers[]** A list of client numbers that authorized for the client service. Each
  client number must also include the country code.
- **services[].commands[].endpoint** The endpoint that will be combined with the base_url to create the complete
  the full path for a given comment.
- **services[].commands[].method** The HTTP method to use for a given endpoint.
- **services[].commands[].pattern** The regex pattern that is used to determine whether to run a command.
- **services[].commands[].args[].datatype** The datatype of the underlying arg. Supported types are as follows:
  - For strings, either "string" or "str" are accepted.
  - For integers, either "integer" or "int" are accepted.
  - For decimals, either "double" or "float" are accepted.
  - For booleans, either "boolean" or "bool" are accepted.
- **services[].commands[].args[].type** The arg class. Supported arg classes are:
  - For query args, use "query".
  - For JSON args, use "json".
  - For endpoint args, use "endpoint".
- **services[].commands[].args[].index** The mapping between a raw arg and the current
  translation. 
- **services[].commands[].args[].path** The JSON/query path to where place/get the arg. For endpoint args, this value
  is ignored and can be removed.
- **services[].commands[].args[].compress_rest** Whether to compress the rest of the input args from the current index
  into an array of the given arg type. 
- **services[].commands[].response.type** The response content type. Supported types are:
  - For JSON response content type, which will enable for further response parsing for success and error cases
    use "json".
  - To simply return the raw response, use "plain_text". 
- **services[].response.success** When type is set to "json", the path to retrieve the response content when
  response status code is 200.
- **services[].response.success** When type is set to "json", the path to retrieve the response content when
  response status code is not 200.

#### **Client Services**
A client service sets up the base service on which specified commands are run against. Services
are defined under the `services:` section in `cot_sm.yaml`. 

The name of the service will service
as the command name. Thus, if the name of the service is `ping` then all underlying commands
will only apply when the `cmd` is `ping`.

#### **Commands**
Commands allow for a programmatic way of hitting different endpoints for a given service. A RESTful
service might contain various endpoints, use different HTTP methods (e.g. GET, POST, ...), and make use of
JSON, and query args for passing data. COT enables one to statically map each positional argument,
starting from the argument right after the command name (the service name).

To add a command to a service, append an entry to `commands` selection of the select service with the
schema:
```yaml
endpoint: "the endpoint of the given service to call for the select command."
method: "The HTTP method to use when calling the endpoint."
pattern: "A regex pattern that is applied to the raw input command in order to identity the command."
response:
- "Schematics to define the response type as logic on how to handle success and erroneous responses."
args:
- "A list of args that perform the mapping of the raw args into corresponding arg groups."
```

### **Encryption Configuration**
The follow environment variables can be defined in the case were encryption is enabled. If
encryption is not enabled, then none of these environment variables need to be set.
- **COT_TEXT_ENCRYPTION=** whether encryption is enabled (true, false)
- **COT_PUBLIC_KEY_FILE=** path to COT's public PGP key
- **COT_PRIVATE_KEY_FILE=** path to COT's private PGP key
- **COT_PASSPHRASE=** passphrase for COT's private PGP key
- **COT_CN_PUBLIC_KEY_DIR=** directory that will store all of the client number public PGP keys
- **COT_SIG_VERIFICATION=** whether signature verification is enabled for PGP
- **COT_BASE64_ENCODING=** whether messages will be base64 encoded

## **Installation**
- Setup GVMS as explained [here](https://github.com/kingcobra2468/GVMS).
- Clone COT and setup [configuration](#configuration).
- Install dependencies with `go get`.
- Launch COT with `go run main.go`.