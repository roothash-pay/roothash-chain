// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "../../interfaces/Itheweb3ChainDepositManager.sol";
import "../../interfaces/IDelegationManager.sol";

abstract contract theweb3ChainDepositManagerStorage is Itheweb3ChainDepositManager {
    bytes32 public constant DOMAIN_TYPEHASH =
        keccak256(
            "EIP712Domain(string name,uint256 chainId,address verifyingContract)"
        );

    bytes32 public constant DEPOSIT_TYPEHASH =
        keccak256(
            "Deposit(address staker,address theweb3ChainBase,uint256 amount,uint256 nonce,uint256 expiry)"
        );

    uint8 internal constant MAX_STAKER_STRATEGY_LIST_LENGTH = 32;

    bytes32 internal _DOMAIN_SEPARATOR;

    uint256 internal withdrawalDelayBlocks;

    IDelegationManager public delegation;

    Itheweb3ChainBase public theweb3ChainBase;

    mapping(address => uint256) public nonces;

    mapping(address => uint256) public stakertheweb3ChainBaseShares;

    mapping(bytes32 => bool) public withdrawalRootPending;

    mapping(address => uint256) internal numWithdrawalsQueued;

    function _inittheweb3ChainDepositManagerStorage(
        IDelegationManager _delegation,
        Itheweb3ChainBase _theweb3ChainBase
    ) internal {
        delegation = _delegation;
        theweb3ChainBase = _theweb3ChainBase;
    }

    uint256[100] private __gap;
}
