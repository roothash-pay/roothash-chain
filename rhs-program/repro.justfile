GIT_COMMIT := ""
GIT_DATE := ""

CANNON_VERSION := "v0.0.0"
OP_PROGRAM_VERSION := "v0.0.0"

GOOS := ""
GOARCH := ""

# Build the cannon binary
cannon:
    #!/bin/bash
    # in devnet scenario, the cannon binary is already built.
    [ -x /app/cannon/bin/cannon ] && exit 0
    cd ../cannon
    make cannon \
        GOOS={{GOOS}} \
        GOARCH={{GOARCH}} \
        GITCOMMIT={{GIT_COMMIT}} \
        GITDATE={{GIT_DATE}} \
        VERSION={{CANNON_VERSION}}

# Build the tw-program-client elf binaries
tw-program-client-mips:
    #!/bin/bash
    cd ../tw-program
    make tw-program-client-mips \
        GOOS=linux \
        GOARCH=mips \
        GOMIPS=softfloat \
        GITCOMMIT={{GIT_COMMIT}} \
        GITDATE={{GIT_DATE}} \
        VERSION={{OP_PROGRAM_VERSION}}

# Run the tw-program-client elf binary directly through cannon's load-elf subcommand.
client TYPE CLIENT_SUFFIX PRESTATE_SUFFIX: cannon tw-program-client-mips
    #!/bin/bash
    /app/cannon/bin/cannon load-elf \
        --type {{TYPE}} \
        --path /app/tw-program/bin/tw-program-client{{CLIENT_SUFFIX}}.elf \
        --out /app/tw-program/bin/prestate{{PRESTATE_SUFFIX}}.bin.gz \
        --meta "/app/tw-program/bin/meta{{PRESTATE_SUFFIX}}.json"

# Generate the prestate proof containing the absolute pre-state hash.
prestate TYPE CLIENT_SUFFIX PRESTATE_SUFFIX: (client TYPE CLIENT_SUFFIX PRESTATE_SUFFIX)
    #!/bin/bash
    /app/cannon/bin/cannon run \
        --proof-at '=0' \
        --stop-at '=1' \
        --input /app/tw-program/bin/prestate{{PRESTATE_SUFFIX}}.bin.gz \
        --meta "" \
        --proof-fmt '/app/tw-program/bin/%d{{PRESTATE_SUFFIX}}.json' \
        --output ""
    mv /app/tw-program/bin/0{{PRESTATE_SUFFIX}}.json /app/tw-program/bin/prestate-proof{{PRESTATE_SUFFIX}}.json

build-default: (prestate "singlethreaded-2" "" "")
build-mt64: (prestate "multithreaded64-3" "64" "-mt64")
build-interop: (prestate "multithreaded64-3" "-interop" "-interop")

build-all: build-default build-mt64 build-interop
