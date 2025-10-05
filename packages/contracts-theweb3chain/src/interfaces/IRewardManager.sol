// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

interface IRewardManager {
    event OperatorAndStakeReward(address chainBase, address operator, uint256 stakerFee, uint256 operatorFee);

    event OperatorClaimReward(address operator, uint256 amount);

    event StakeHolderClaimReward(address stakeHolder, address chainBase, uint256 amount);

    function payFee(address chainBase, address operator, uint256 baseFee) external;
    function operatorClaimReward() external returns (bool);
    function getStakeHolderAmount(address chainBase) external returns (uint256);
    function stakeHolderClaimReward(address chainBase) external returns (bool);
    function updateStakePercent(uint256 _stakePercent) external;
}
