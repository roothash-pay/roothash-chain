<div align="center">
  <br />
  <br />
  <a href="https://cpchain.io"><img alt="CpChain" src="./docs/assets/cpchain.svg" width=600></a>
  <br />
  <h3><a href="https://cpchain.io">CpChain</a> is Ethereum, scaled.</h3>
  <br />
</div>

**Table of Contents**

<!--TOC-->

- [What is CpChain?](#what-is-cpchain)
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

## What is CpChain?

[CpChain](https://www.cpchain.io/) is a project dedicated to scaling Ethereum's technology and expanding its ability to coordinate people from across the world to build effective decentralized economies and governance systems. The [CpChain Collective](https://www.cpchain.io/vision) builds open-source software that powers scalable blockchains and aims to address key governance and economic challenges in the wider Ethereum ecosystem. CpChain operates on the principle of **impact=profit**, the idea that individuals who positively impact the Collective should be proportionally rewarded with profit. **Change the incentives and you change the world.**

## Documentation

- If you want to build on top of CpChain Mainnet, refer to the [CpChain Documentation](https://docs.cpchain.io)
- If you want to build your own chain based CpChain, refer to the [CpChain Guide](https://docs.cpchain.io/stack/getting-started) and make sure to understand this repository's [Development and Release Process](#development-and-release-process)

## Specification

Detailed specifications for the CpChain can be found within the [CpChain Specs](https://github.com/cpchain-network/specs) repository.

## Community

General discussion happens most frequently on the [CpChain discord](https://discord.gg/cpchain).
Governance discussion can also be found on the [CpChain Governance Forum](https://gov.cpchain.io/).

## Contributing

The CpChain is a collaborative project. By collaborating on free, open software and shared standards, the CpChain Collective aims to prevent siloed software development and rapidly accelerate the development of the Ethereum ecosystem. Come contribute, build the future, and redefine power, together.

[CONTRIBUTING.md](./CONTRIBUTING.md) contains a detailed explanation of the contributing process for this repository. Make sure to use the [Developer Quick Start](./CONTRIBUTING.md#development-quick-start) to properly set up your development environment.

[Good First Issues](https://github.com/cpchain-network/cpchain/issues?q=is:open+is:issue+label:D-good-first-issue) are a great place to look for tasks to tackle if you're not sure where to start, and see [CONTRIBUTING.md](./CONTRIBUTING.md) for info on larger projects.

## Security Policy and Vulnerability Reporting

Please refer to the canonical [Security Policy](https://github.com/cpchain-network/.github/blob/master/SECURITY.md) document for detailed information about how to report vulnerabilities in this codebase.
Bounty hunters are encouraged to check out the [CpChain Immunefi bug bounty program](https://immunefi.com/bounty/cpchain/).
The CpChain Immunefi program offers up to $2,000,042 for in-scope critical vulnerabilities.

## Directory Structure

<pre>
├── <a href="./docs">docs</a>: A collection of documents including audits and post-mortems
├── <a href="./kurtosis-devnet">kurtosis-devnet</a>: OP-Stack Kurtosis devnet
├── <a href="./op-chain-ops">op-chain-ops</a>: State surgery utilities
├── <a href="./op-e2e">op-e2e</a>: End-to-End testing of all cpchain components in Go
├── <a href="./op-node">op-node</a>: consensus-layer of CpChain
├── <a href="./op-service">op-service</a>: Common codebase utilities
├── <a href="./ops">ops</a>: Various operational packages
├── <a href="./packages">packages</a>
│   ├── <a href="./packages/contracts-cpchain">contracts-cpchain</a>: CpChain smart contracts
├── <a href="./.semgrep">semgrep</a>: Semgrep rules and tests
</pre>

## Development and Release Process

### Overview

Please read this section carefully if you're planning to fork or make frequent PRs into this repository.

### Production Releases

TBD

### Development branch

TBD

## License

All other files within this repository are licensed under the [MIT License](https://github.com/cpchain-network/cpchain/blob/master/LICENSE) unless stated otherwise.
