# sensu-irc-handler

This is a [Sensu Event Handler](https://docs.sensu.io/sensu-go/5.2/reference/handlers/#how-do-sensu-handlers-work)
that sends event data to a configured IRC room. 

## Installation 
The latest release can be downloaded via the [Releases page](https://github.com/sensu-utils/sensu-irc-handler/releases)

You can also download the latest asset form [Bonsai](https://bonsai.sensu.io/assets/sensu-utils/sensu-irc-handler)

## Configuration
The following is an example configuration for the IRC handler:

1. Create a local file named `irc-handler.json` with the following contents.
    ```json
    {
        "api_version": "core/v2",
        "type": "Handler",
        "metadata": {
            "namespace": "default",
            "name": "irc"
        },
        "spec": {
            "type": "pipe",
            "command": "sensu-irc-handler",
            "timeout": 30,
            "filters": [
                "is_incident"
            ]
        }
    }
    ```
2. Run the command `sensuctl create -f irc-handler.json`
3. In your [check definition](https://docs.sensu.io/sensu-go/5.2/reference/checks/#spec-attributes) configure `handlers` to contain `irc`.
 
`## Usage Examples