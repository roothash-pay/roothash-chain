// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import {Itheweb3ChainBase} from "./Itheweb3ChainBase.sol";

interface Itheweb3ChainDepositManager {
    event Deposit(address staker, Itheweb3ChainBase chainbase, uint256 shares);

    event UpdatedThirdPartyTransfersForbidden(Itheweb3ChainBase chainbase, bool value);

    event theweb3ChainBaseWhitelisterChanged(address previousAddress, address newAddress);

    event theweb3ChainBaseAddedToDepositWhitelist(Itheweb3ChainBase chainbase);

    event theweb3ChainBaseRemovedFromDepositWhitelist(Itheweb3ChainBase chainbase);

    function depositIntotheweb3Chain(uint256 amount) external payable returns (uint256 shares);

    function depositIntotheweb3ChainWithSignature(uint256 amount, address staker, uint256 expiry, bytes memory signature) external payable returns (uint256 shares);

    function removeShares(address staker, uint256 shares) external;

    function addShares(address staker,  uint256 shares) external;

    function withdrawSharesAsCp(address recipient,  uint256 shares) external;

    function stakertheweb3ChainBaseShares(address staker) external view returns (uint256 shares);

    function getDeposits(address staker) external view returns (uint256);
}
