// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "@openzeppelin-upgrades/contracts/utils/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import "../../interfaces/IRewardManager.sol";
import "../../interfaces/IDelegationManager.sol";
import "../../interfaces/ICpChainDepositManager.sol";


abstract contract RewardManagerStorage is Initializable, OwnableUpgradeable, ReentrancyGuardUpgradeable, IRewardManager {
    using SafeERC20 for IERC20;

    IDelegationManager public immutable delegationManager;

    ICpChainDepositManager public immutable cpChainDepositManager;

    IERC20 public immutable rewardTokenAddress;

    uint256 public stakePercent;

    address public rewardManager;

    address public payFeeManager;

    mapping(address => uint256) public chainBaseStakeRewards;
    mapping(address => uint256) public operatorRewards;

    constructor(IDelegationManager _delegationManager, ICpChainDepositManager _cpChainDepositManager, IERC20 _rewardTokenAddress) {
        delegationManager = _delegationManager;
        cpChainDepositManager = _cpChainDepositManager;
        rewardTokenAddress = _rewardTokenAddress;
    }

    uint256[100] private __gap;
}
