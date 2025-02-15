# Plugins

Ika's plugin architecture provides the building blocks for extending gateway functionality. While the core focuses on routing, plugins add the features you need.

## Core Plugins

Core plugins are maintained by the Ika team and provide essential gateway functionality. Each plugin undergoes:

- Thorough testing
- Regular maintenance
- Security updates
- Detailed documentation

### Access Log (`access-log`)

HTTP request logging with configurable formats.
[Learn more →](/plugins/access-log)

### Basic Auth (`basic-auth`)

HTTP Basic Authentication support.
[Learn more →](/plugins/basic-auth)

### Request Modifier (`req-modifier`)

Request transformation and adaptation.
[Learn more →](/plugins/request-modifier)

### Fail2Ban (`fail2ban`)

IP-based threat protection.
[Learn more →](/plugins/fail2ban)

## Plugin Configuration

Plugins can be configured at multiple levels to control gateway behavior.

### Namespace-Level Configuration

TODO

### Route-Level Configuration

TODO

## Next Steps

- [Access Log Plugin](/plugins/access-log) - Logging configuration
- [Basic Auth Plugin](/plugins/basic-auth) - Authentication setup
- [Request Modifier Plugin](/plugins/request-modifier) - Request transformation
- [Fail2Ban Plugin](/plugins/fail2ban) - Security settings
