// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "../../interfaces/ICpChainDepositManager.sol";
import "../../interfaces/IDelegationManager.sol";

abstract contract CpChainDepositManagerStorage is ICpChainDepositManager {
    bytes32 public constant DOMAIN_TYPEHASH = keccak256("EIP712Domain(string name,uint256 chainId,address verifyingContract)");

    bytes32 public constant DEPOSIT_TYPEHASH = keccak256("Deposit(address staker,address cpChainBase,uint256 amount,uint256 nonce,uint256 expiry)");

    uint8 internal constant MAX_STAKER_STRATEGY_LIST_LENGTH = 32;

    IDelegationManager public immutable delegation;

    bytes32 internal _DOMAIN_SEPARATOR;

    mapping(address => uint256) public nonces;

    address public cpChainWhitelister;

    uint256 internal withdrawalDelayBlocks;

    mapping(address => mapping(ICpChainBase => uint256)) public stakerCpChainBaseShares;

    mapping(address => ICpChainBase[]) public stakerCpChainBaseList;

    mapping(bytes32 => bool) public withdrawalRootPending;

    mapping(address => uint256) internal numWithdrawalsQueued;

    mapping(ICpChainBase => bool) public cpChainWhitelistedForDeposit;

    mapping(address => uint256) internal beaconChainETHSharesToDecrementOnWithdrawal;


    mapping(ICpChainBase => bool) public thirdPartyTransfersForbidden;


    constructor(IDelegationManager _delegation) {
        delegation = _delegation;
    }


    uint256[100] private __gap;
}



