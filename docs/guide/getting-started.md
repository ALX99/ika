# Getting Started

Deploy your first API Gateway in 60 seconds.

## Installation

Ika is distributed as a lightweight single binary. Choose your preferred installation method:

::: code-group

```bash [💻 Source]
# Install
go install github.com/alx99/ika/cmd/ika@latest

# Verify
ika -help
```

```bash [🐳 Docker]
# Install
docker pull ghcr.io/alx99/ika:latest

# Verify
docker run -it --rm ghcr.io/alx99/ika:latest -help
```

```bash [📦 Binary]
# Download and extract
VERSION="v0.0.6"
URL="https://github.com/ALX99/ika/releases/download/$VERSION/ika_Linux_$(uname -m).tar.gz"
curl -Lo- "$URL" | tar -xz --wildcards 'ika'

# Verify
./ika -help
```

:::

## Basic Configuration

Create a file named `ika.yaml` with this minimal configuration:

<<<@/examples/configs/basic.yaml

### Understanding Namespaces

Namespaces in Ika are isolated routing groups. They help you:

- Organize related routes together
- Apply plugins to specific route groups
- Maintain isolation between different parts of your gateway

For example, if you have two namespaces using the same rate limit plugin:

- Each namespace has its own independent rate limit
- A plugin failure in one namespace won't affect the other
- Configuration changes in one namespace don't impact others

## Running Ika

Start Ika with your configuration:

```bash
ika -config ika.yaml
```

::: tip
By default, Ika looks for `ika.yaml` in the current directory. Use `-config` to specify a different path.
:::

If successful, you'll see output like:

```log{5}
Feb 15 13:51:13.869 INF Logger initialized config.level=debug config.format=text config.flushInterval=0s config.addSource=false
Feb 15 13:51:13.870 DBG Log buffering disabled
Feb 15 13:51:13.870 INF Building router namespaceCount=1
Feb 15 13:51:13.870 DBG Built namespace ns=myFirstNamespace dur=340.928µs
Feb 15 13:51:14.871 INF Ika has started startupTime=1.002s version="" goVersion="go1.24.0 X:synctest"
```

## Testing Your Gateway

Try visiting [http://localhost:8888](http://localhost:8888). Nothing happens? That's expected!

Remember: Ika starts with minimal functionality. However, the configuration above can proxy requests to any website.
Can you figure it out?

::: code-group

```bash [Challenge]
# Try to figure out how to make a request to httpbin.org through your gateway!
# Hint: Look at the HOST header and connection settings
```

```bash [Solution]
# Use curl's connect-to feature to route through your gateway
HOST="httpbin.org"
curl --connect-to "localhost:8888:$HOST:443" "https://$HOST/get"

# You should see a JSON response from httpbin.org!
```

:::

## Adding Real Functionality

Let's make our gateway more useful by adding some basic features:

TODO

## Next Steps

Now that you have a basic gateway running, you can:

- Learn about [Plugins](/plugins/) to add more functionality
- Explore advanced [Configuration](/guide/configurationTODO) options
- Check out more [Examples](/examplesTODO) of real-world setups
