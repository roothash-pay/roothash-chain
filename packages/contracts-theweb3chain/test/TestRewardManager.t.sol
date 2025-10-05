// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@/core/pos/theweb3ChainBase.sol";
import "@/core/pos/theweb3ChainDepositManager.sol";
import "@/core/pos/DelegationManager.sol";
import "@/core/pos/SlashingManager.sol";
import "@/core/pos/RewardManager.sol";

import "@/access/PauserRegistry.sol";
import "@/interfaces/IDelegationManager.sol";
import "@/interfaces/ISlashingManager.sol";
import "@/interfaces/ISignatureUtils.sol";
import "@/interfaces/IRewardManager.sol";

import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract RewardManagerTest is Test {
    theweb3ChainBase public theweb3ChainBase;
    PauserRegistry public pauserregistry;
    theweb3ChainDepositManager public theweb3ChainDepositManager;
    DelegationManager public delegationManager;
    SlashingManager public slashingManager;
    RewardManager public rewardManager;
    IDelegationManager.QueuedWithdrawalParams[] public params;
    IDelegationManager.Withdrawal[] public withdraws;

    address public user1 = address(0x01);
    address public user2 = address(0x11);
    address public pauser1 = address(0x02);
    address public pauser2 = address(0x03);
    address[] public pausers;
    address public unpauser = address(0x04);

    address public owner = address(0x05);
    address public operator = address(0x06);
    address public operator1 = address(0x07);
    address public rewardingmanager = address(0x08);
    address public payFeeManager = address(0x09);
    address public slasher = address(0x10);
    address public slashingReceipt = address(100);

    function setUp() public {
        pausers.push(pauser1);
        pausers.push(pauser2);
        pauserregistry = new PauserRegistry(pausers, unpauser);

        theweb3ChainBase logic1 = new theweb3ChainBase();
        TransparentUpgradeableProxy proxy1 = new TransparentUpgradeableProxy(address(logic1), owner, "");

        theweb3ChainDepositManager logic2 = new theweb3ChainDepositManager();
        TransparentUpgradeableProxy proxy2 = new TransparentUpgradeableProxy(address(logic2), owner, "");

        DelegationManager logic3 = new DelegationManager();
        TransparentUpgradeableProxy proxy3 = new TransparentUpgradeableProxy(address(logic3), owner, "");

        SlashingManager logic4 = new SlashingManager();
        TransparentUpgradeableProxy proxy4 = new TransparentUpgradeableProxy(address(logic4), owner, "");

        RewardManager logic5 = new RewardManager();
        TransparentUpgradeableProxy proxy5 = new TransparentUpgradeableProxy(address(logic5), owner, "");

        theweb3ChainBase = theweb3ChainBase(payable(address(proxy1)));
        theweb3ChainDepositManager = theweb3ChainDepositManager(payable(address(proxy2)));
        delegationManager = DelegationManager(payable(address(proxy3)));
        slashingManager = SlashingManager(payable(address(proxy4)));
        rewardManager = RewardManager(payable(address(proxy5)));

        theweb3ChainBase.initialize(
            pauserregistry, 1 ether, 10 ether, Itheweb3ChainDepositManager(address(theweb3ChainDepositManager))
        );
        theweb3ChainDepositManager.initialize(owner, IDelegationManager(address(delegationManager)), theweb3ChainBase);
        delegationManager.initialize(
            owner,
            IPauserRegistry(address(pauserregistry)),
            0,
            0,
            theweb3ChainDepositManager,
            theweb3ChainBase,
            slashingManager
        );
        slashingManager.initialize(
            owner, IDelegationManager(address(delegationManager)), slasher, 0.1 ether, slashingReceipt
        );
        rewardManager.initialize(
            owner,
            rewardingmanager,
            payFeeManager,
            50,
            IPauserRegistry(address(pauserregistry)),
            delegationManager,
            theweb3ChainDepositManager
        );

        vm.deal(user1, 100 ether);
        vm.deal(user2, 100 ether);
        vm.deal(address(rewardManager), 500 ether);
        _registerAndDelegate();
    }

    function _registerAndDelegate() internal {
        DelegationManager.OperatorDetails memory od = IDelegationManager.OperatorDetails({
            earningsReceiver: operator,
            delegationApprover: address(0),
            stakerOptOutWindowBlocks: 100
        });
        DelegationManager.OperatorDetails memory od1 = IDelegationManager.OperatorDetails({
            earningsReceiver: operator1,
            delegationApprover: address(0),
            stakerOptOutWindowBlocks: 100
        });
        ISignatureUtils.SignatureWithExpiry memory emptySignatureAndExpiry;

        vm.prank(operator);
        delegationManager.registerAsOperator(od, "node-url");
        vm.prank(operator1);
        delegationManager.registerAsOperator(od1, "node-url");

        vm.prank(user1);
        theweb3ChainDepositManager.depositIntotheweb3Chain{value: 2 ether}(2 ether);
        vm.prank(user2);
        theweb3ChainDepositManager.depositIntotheweb3Chain{value: 2 ether}(2 ether);

        vm.prank(user1);
        delegationManager.delegateTo(operator, emptySignatureAndExpiry, bytes32(0));
        vm.prank(user2);
        delegationManager.delegateTo(operator1, emptySignatureAndExpiry, bytes32(0));

        assertEq(delegationManager.stakerDelegateSharesToOperator(operator, user1), 2 ether);
        assertEq(delegationManager.stakerDelegateSharesToOperator(operator1, user2), 2 ether);
    }

    function testPayFeeAndClaimRewards() public {
        vm.prank(payFeeManager);
        rewardManager.payFee(address(theweb3ChainBase), operator, 400 ether);

        assertEq(rewardManager.operatorRewards(operator), 100 ether);
        assertEq(rewardManager.operatorRewards(operator1), 0);

        vm.deal(address(rewardManager), 500 ether);

        vm.prank(operator);
        bool success = rewardManager.operatorClaimReward();
        assertTrue(success);
        assertEq(operator.balance, 100 ether);

        vm.prank(user1);
        bool success1 = rewardManager.stakeHolderClaimReward(address(theweb3ChainBase));
        assertTrue(success1);
        assertEq(user1.balance, 98 ether + 50 ether);

        vm.prank(user2);
        bool success2 = rewardManager.stakeHolderClaimReward(address(theweb3ChainBase));
        assertTrue(success2);

        assertEq(user2.balance, 98 ether + 50 ether);
    }

    function testPayFeeRevertsIfZeroShares() public {
        vm.prank(payFeeManager);
        vm.expectRevert("RewardManager payFee: one of totalShares and operatorShares is zero");
        rewardManager.payFee(address(theweb3ChainBase), slasher, 100 ether);
    }

    function testStakeHolderClaimRewardWorks() public {
        vm.prank(payFeeManager);
        rewardManager.payFee(address(theweb3ChainBase), operator, 100 ether);

        vm.prank(slasher);
        vm.expectRevert("RewardManager operatorClaimReward: stake holder amount need more then zero");
        rewardManager.stakeHolderClaimReward(address(theweb3ChainBase));

        vm.deal(address(rewardManager), 0.1 ether);

        vm.prank(user1);
        vm.expectRevert("RewardManager operatorClaimReward: Reward Token balance insufficient");
        rewardManager.stakeHolderClaimReward(address(theweb3ChainBase));

        vm.deal(address(rewardManager), 500 ether);

        vm.prank(user1);
        bool success = rewardManager.stakeHolderClaimReward(address(theweb3ChainBase));
        assertTrue(success);
        assertEq(user1.balance, 110.5 ether);
    }

    function testUpdateStakePercent() public {
        vm.expectRevert("RewardManager.only reward manager can call this function");
        rewardManager.updateStakePercent(80);

        vm.prank(rewardingmanager);
        rewardManager.updateStakePercent(80);
        assertEq(rewardManager.stakePercent(), 80);
    }

    function testOperatorClaimRewardFailsWhenBalanceInsufficient() public {
        vm.prank(slasher);
        vm.expectRevert("RewardManager.only pay fee manager can call this function");
        rewardManager.payFee(address(theweb3ChainBase), operator, 400 ether);

        vm.prank(payFeeManager);
        rewardManager.payFee(address(theweb3ChainBase), operator, 400 ether);

        vm.prank(slasher);
        vm.expectRevert("RewardManager operatorClaimReward: operator claim amount need more then zero");
        rewardManager.operatorClaimReward();

        vm.prank(operator);
        vm.deal(address(rewardManager), 0.1 ether);
        vm.expectRevert("RewardManager operatorClaimReward: Reward Token balance insufficient");
        rewardManager.operatorClaimReward();

        vm.deal(address(rewardManager), 500 ether);
        vm.prank(operator);
        rewardManager.operatorClaimReward();

        assertEq(operator.balance, 100 ether);
    }
}
