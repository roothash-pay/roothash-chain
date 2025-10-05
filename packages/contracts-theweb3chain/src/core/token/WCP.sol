// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {WrappedTW} from "src/universal/WrappedTW.sol";

contract WTW is WrappedTW {
    string public constant version = "0.0.1";

    function name() external pure override returns (string memory name_) {
        name_ = "Wrapped TW";
    }

    function symbol() external pure override returns (string memory symbol_) {
        symbol_ = string.concat("WTW");
    }
}
