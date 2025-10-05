// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {FeeVault} from "src/core/fee/FeeVault.sol";
import {Types} from "src/libraries/Types.sol";

contract SequencerFeeVault is FeeVault {
    /// @custom:semver 0.0.1
    string public constant version = "0.0.1";

    constructor(address _recipient, uint256 _minWithdrawalAmount) FeeVault(_recipient, _minWithdrawalAmount) {}

    function feeWallet() public view returns (address) {
        return RECIPIENT;
    }
}
