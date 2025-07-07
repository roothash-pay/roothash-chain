// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import {ICpChainBase} from "./ICpChainBase.sol";

interface ICpChainDepositManager {
    event Deposit(address staker, ICpChainBase chainbase, uint256 shares);

    event UpdatedThirdPartyTransfersForbidden(ICpChainBase chainbase, bool value);

    event CpChainBaseWhitelisterChanged(address previousAddress, address newAddress);

    event CpChainBaseAddedToDepositWhitelist(ICpChainBase chainbase);

    event CpChainBaseRemovedFromDepositWhitelist(ICpChainBase chainbase);

    function depositIntoCpChain(ICpChainBase chainbase, uint256 amount) external payable returns (uint256 shares);

    function depositIntoCpChainWithSignature(
        ICpChainBase chainbase,
        uint256 amount,
        address staker,
        uint256 expiry,
        bytes memory signature
    ) external payable returns (uint256 shares);

    function removeShares(address staker, ICpChainBase chainbase, uint256 shares) external;

    function addShares(address staker, ICpChainBase chainbase, uint256 shares) external;

    function withdrawSharesAsCp(address recipient, ICpChainBase chainbase, uint256 shares) external;

    function stakerCpChainBaseShares(address user, ICpChainBase chainbase) external view returns (uint256 shares);

    function getDeposits(address staker) external view returns (ICpChainBase[] memory, uint256[] memory);

    function stakerCpChainBaseListLength(address staker) external view returns (uint256);

    function addStrategiesToDepositWhitelist(
        ICpChainBase[] calldata strategiesToWhitelist,
        bool[] calldata thirdPartyTransfersForbiddenValues
    ) external;

    function removeStrategiesFromDepositWhitelist(ICpChainBase[] calldata strategiesToRemoveFromWhitelist) external;

    function cpChainWhitelister() external view returns (address);

    function thirdPartyTransfersForbidden(ICpChainBase chainbase) external view returns (bool);
}
