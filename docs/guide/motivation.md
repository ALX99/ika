# Motivation

Ika was born out of real-world frustration with existing API Gateways in the Go ecosystem. While many excellent gateways exist, we found that their approach to feature completeness often creates more problems than it solves.

## The Problem with Feature-Complete Gateways

Most API Gateways, in an attempt to be "complete" and attract users, come preloaded with a vast array of features. While this might sound appealing at first, it leads to several significant issues:

### 1. Feature Bloat

- Most built-in features go unused in production
- Each feature adds configuration complexity
- Users pay the cost (complexity, performance, maintenance) for features they don't use

### 2. Hidden Complexity

Features aren't free - they come with:

- Increased codebase complexity
- More potential for bugs
- Harder to understand behavior
- Complex interaction between features

### 3. Deployment Overhead

More features mean:

- Larger binary sizes
- Slower container deployments
- Higher memory usage at runtime
- Longer cold starts in serverless environments

### 4. Limited Customization

When you actually need custom functionality:

- You're constrained by the gateway's extension points
- Custom features must work around existing ones
- Plugin APIs are often an afterthought
- Integration with existing features is challenging

## The Birth of Ika

These challenges led us to question: What if we took the opposite approach? Instead of starting with everything, what if we started with nothing but the essentials?

This thought experiment led to Ika's core principles:

1. Start with routing - the one thing every API Gateway must do
2. Make the plugin system first-class
3. Keep the core minimal and focused
4. Give users the same power as core developers

The result is a gateway that:

- Has minimal core dependencies
- Lets you add exactly what you need
- Makes extending functionality natural
- Keeps complexity proportional to your needs

::: tip
Want to see how minimal we are? Check our [go.mod](https://github.com/alx99/ika/blob/main/go.mod) file - it's remarkably small for an API Gateway!
:::

## Further Reading

- [Why Ika](/guide/why-ika) - Overview of Ika's approach
- [Getting Started](/guide/getting-started) - Try Ika for yourself
- [Plugins](/plugins/) - Learn about extending Ika
