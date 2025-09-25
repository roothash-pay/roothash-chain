<div align="center">
  <br />
  <br />
  <a href="https://theweb3Chain.com"><img alt="theweb3Chain" src="./docs/assets/theweb3Chain.svg" width=600></a>
  <br />
  <h3><a href="https://theweb3Chain.com">theweb3Chain</a> is Ethereum, scaled.</h3>
  <br />
</div>

**Table of Contents**

<!--TOC-->

- [What is theweb3Chain?](#what-is-theweb3Chain)
- [Documentation](#documentation)
- [Specification](#specification)
- [Community](#community)
- [Contributing](#contributing)
- [Security Policy and Vulnerability Reporting](#security-policy-and-vulnerability-reporting)
- [Directory Structure](#directory-structure)
- [Development and Release Process](#development-and-release-process)
  - [Overview](#overview)
  - [Production Releases](#production-releases)
  - [Development branch](#development-branch)
- [License](#license)

<!--TOC-->

## What is theweb3Chain?

[theweb3Chain](https://www.theweb3Chain.com/) is a project dedicated to scaling Ethereum's technology and expanding its ability to coordinate people from across the world to build effective decentralized economies and governance systems. The [theweb3Chain Collective](https://www.theweb3Chain.com/vision) builds open-source software that powers scalable blockchains and aims to address key governance and economic challenges in the wider Ethereum ecosystem. theweb3Chain operates on the principle of **impact=profit**, the idea that individuals who positively impact the Collective should be proportionally rewarded with profit. **Change the incentives and you change the world.**

## Documentation

- If you want to build on top of theweb3Chain Mainnet, refer to the [theweb3Chain Documentation](https://docs.theweb3Chain.com)
- If you want to build your own chain based theweb3Chain, refer to the [theweb3Chain Guide](https://docs.theweb3Chain.com/stack/getting-started) and make sure to understand this repository's [Development and Release Process](#development-and-release-process)

## Specification

Detailed specifications for the theweb3Chain can be found within the [theweb3Chain Specs](https://github.com/theweb3Chain-network/specs) repository.

## Community

General discussion happens most frequently on the [theweb3Chain discord](https://discord.gg/theweb3Chain).
Governance discussion can also be found on the [theweb3Chain Governance Forum](https://gov.theweb3Chain.com/).

## Contributing

The theweb3Chain is a collaborative project. By collaborating on free, open software and shared standards, the theweb3Chain Collective aims to prevent siloed software development and rapidly accelerate the development of the Ethereum ecosystem. Come contribute, build the future, and redefine power, together.

[CONTRIBUTING.md](./CONTRIBUTING.md) contains a detailed explanation of the contributing process for this repository. Make sure to use the [Developer Quick Start](./CONTRIBUTING.md#development-quick-start) to properly set up your development environment.

[Good First Issues](https://github.com/theweb3Chain-network/theweb3Chain/issues?q=is:open+is:issue+label:D-good-first-issue) are a great place to look for tasks to tackle if you're not sure where to start, and see [CONTRIBUTING.md](./CONTRIBUTING.md) for info on larger projects.

## Security Policy and Vulnerability Reporting

Please refer to the canonical [Security Policy](https://github.com/theweb3Chain-network/.github/blob/master/SECURITY.md) document for detailed information about how to report vulnerabilities in this codebase.
Bounty hunters are encouraged to check out the [theweb3Chain Immunefi bug bounty program](https://immunefi.com/bounty/theweb3Chain/).

## Directory Structure

<pre>
├── <a href="./docs">docs</a>: A collection of documents including audits and post-mortems
├── <a href="./theweb3-chain-ops">theweb3-chain-ops</a>: State surgery utilities
├── <a href="./tw-node">tw-node</a>: consensus-layer of theweb3Chain
├── <a href="./tw-service">tw-service</a>: Common codebase utilities
├── <a href="./ops">ops</a>: Various operational packages
├── <a href="./packages">packages</a>
│   ├── <a href="./packages/contracts-theweb3Chain">contracts-theweb3Chain</a>: theweb3Chain smart contracts
</pre>

## Development and Release Process

### Overview

Please read this section carefully if you're planning to fork or make frequent PRs into this repository.

### Production Releases

TBD

### Development branch

TBD

## License

All other files within this repository are licensed under the [MIT License](https://github.com/theweb3Chain-network/theweb3Chain/blob/master/LICENSE) unless stated otherwise.
