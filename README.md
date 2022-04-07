# **COT**
Commands over text(COT) is generic runtime for creating commands from user-built
services.

## **Terminology**
- **Client Number=** phone number that sends the command requests.
- **GVoice Number=** Gvoice phone number that services as receiver for commands that
  came from client numbers.
- **GVMS=** microservice for interacting with GVoice's APIs.
- **Client Service=** user-defined service that handles the command and returns
  the response to COT.

## **Architecture**
![photo](images/cot.jpg)

COT, being only a generic service, enables one to create any command they want
as long as one defines it within `cot_sm.yaml`. Within the config file,
one would specify:
1. Name of the command.
2. User service that will execute the command.
3. List of client numbers authorized to run the command.

### **Flow**
The end-to-end flow starting from the client number is as follows:
1. Client sends command in "[cmd] [arg 1] [arg 2] ... [arg N]" format to the configured
   GVoice number.
3. COT would have initialized a worker that checks the (GVoice Number, Client Number)
   link periodically via GVMS. By listening to only the subnet of defined Client Numbers,
   COT will by nature ignore all numbers that have not been whitelisted within `cot_sm.yaml`
   by any of the services.
4. COT parsers the command and checks if the client number is authorized to run this command.
   Non-authorized commands will be rejected.
5. The arguments of the command will be sent as an array of args to the client service's endpoint
   that was defined for that specified command within `cot_sm.yaml`.
6. The output of the command would then be transmitted back to the client number.

#### **Example**
Assuming these commands are the only commands defined within `cot_sm.yaml`
```yaml
services:
  - name: lights
    base_uri: "http://localhost:9877"
    client_numbers:
    - 1415111111
  - name: email
    base_uri: "http://localhost:9876"
    client_numbers:
    - 1415111111
    - 1415111112
```

If `1415111111` sends a command such as "lights off", COT will verify that this number is
authorized and will send `{args: "off"}` to `/cmd` endpoint of `http://localhost:9877`.
However, if `1415111112` tries to do the same, the request will fail as it is not authorized.
Likewise, another number `1415111113` will fail regardless of the `lights` or `email` command as
they are not authorized for either. All 3 numbers will be rejected for any other commands as
no other commands exist.

## **Defining User Commands & Services**
All commands are defined as list items under the `services:` section. Each command must follow
this schematic
```yaml
- name: "name of command"
  base_uri: "base uri of server executing the command"
  client_numbers:
  - "client number 1"
  - "...."
  - "client number n"
```

As of now, a client service must expose the `/cmd` endpoint for the POST method.
Arguments will be passed in as JSON as `{args: [...]}` via an arg array in the exact same
order they were sent by the client number. The endpoint must return a response with a message
defined in the `message` key. Optionally, an error can be returned via the `error` key
```json
{
    "message": "",
    "error": "optional"
}
```

## **Configuration**
Configuration is done via the `cot_sm.yaml` file which needs to be copied/renamed from 
`cot_sm_template.yaml`. Aside from defining services which were explained
[here](#defining-user-commands), the gvoice number that all client numbers will be sending
commands to needs to be defined via `gvoice_number`. Similarly, GVMS config needs to be defined
which sets the binding of what hostname and port GVMS is running on.

## **Installation**
- Setup GVMS as explained [here](https://github.com/kingcobra2468/GVMS).
- Clone COT and setup [configuration](#configuration).
- Install dependencies with `go get`.
- Launch COT with `go run main.go`.