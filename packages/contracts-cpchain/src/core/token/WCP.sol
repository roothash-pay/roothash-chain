// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { WrappedCP  } from "src/universal/WrappedCP.sol";

contract WCP is WrappedCP {
    string public constant version = "0.0.1";

    function name() external pure override returns (string memory name_) {
        name_ = "Wrapped CP";
    }

    function symbol() external pure override returns (string memory symbol_) {
        symbol_ = string.concat("WCP");
    }
}
