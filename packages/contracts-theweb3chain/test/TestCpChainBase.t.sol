// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@/core/pos/theweb3ChainBase.sol";
import "@/access/PauserRegistry.sol";

import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract theweb3ChainBaseTest is Test {
    theweb3ChainBase public theweb3ChainBase;
    PauserRegistry public pauserregistry;

    address public user1 = address(0x01);
    address public pauser1 = address(0x02);
    address public pauser2 = address(0x03);
    address[] public pausers;
    address public unpauser = address(0x04);
    address public owner = address(0x05);
    address public strategyManager = address(0x06);

    function setUp() public {
        vm.deal(strategyManager, 10 ether);
        pausers.push(pauser1);
        pausers.push(pauser2);
        pauserregistry = new PauserRegistry(pausers, unpauser);

        theweb3ChainBase logic = new theweb3ChainBase();
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(address(logic), owner, "");

        theweb3ChainBase = theweb3ChainBase(payable(address(proxy)));

        theweb3ChainBase.initialize(
            IPauserRegistry(address(pauserregistry)),
            1 ether,
            10 ether,
            Itheweb3ChainDepositManager(address(strategyManager))
        );
    }

    function testDepositShouldMintShares() public {
        uint256 balanceBefore = address(theweb3ChainBase).balance;

        vm.deal(strategyManager, 10 ether);
        vm.prank(strategyManager);

        uint256 newShares = theweb3ChainBase.deposit{value: 10 ether}(10 ether, user1);

        uint256 balanceAfter = address(theweb3ChainBase).balance;

        assertGt(newShares, 0);
        assert(balanceAfter > balanceBefore);
        assertEq(theweb3ChainBase.totalShares(), newShares);
    }

    function testDepositTooLowOrHighShouldRevert() public {
        vm.expectRevert("theweb3ChainBase: deposit token must more than min deposit amount");
        vm.deal(strategyManager, 10 ether);
        vm.prank(strategyManager);

        theweb3ChainBase.deposit{value: 0.5 ether}(0.5 ether, user1);

        vm.expectRevert("theweb3ChainBase: deposit token must less than max deposit amount");
        vm.deal(strategyManager, 15 ether);
        vm.prank(strategyManager);

        theweb3ChainBase.deposit{value: 12 ether}(12 ether, user1);
    }

    function testWithdrawShouldSendEth() public {
        vm.prank(strategyManager);

        uint256 newShares = theweb3ChainBase.deposit{value: 5 ether}(5 ether, user1);

        uint256 balanceBefore = address(theweb3ChainBase).balance;

        vm.prank(user1);
        vm.expectRevert("theweb3ChainBase.onlyStrategyManager");
        theweb3ChainBase.withdraw(user1, newShares);

        vm.prank(strategyManager);
        theweb3ChainBase.withdraw(user1, newShares);
        uint256 balanceAfter = address(theweb3ChainBase).balance;

        assertGt(balanceBefore, balanceAfter);
        assert(user1.balance == 5 ether);
    }

    function testWithdrawTooMuchShouldRevert() public {
        vm.expectRevert("theweb3ChainBase.withdraw: amountShares must be less than or equal to totalShares");
        vm.prank(strategyManager);
        theweb3ChainBase.withdraw(address(this), 100 ether);
    }

    function testSetDepositLimits() public {
        vm.prank(strategyManager);
        theweb3ChainBase.setDepositLimits(2 ether, 200 ether);
        (uint256 minD, uint256 maxD) = theweb3ChainBase.getDepositLimits();
        assertEq(minD, 2 ether);
        assertEq(maxD, 200 ether);
    }

    function testSharesToUnderlyingMutableView() public view {
        uint256 underlying = theweb3ChainBase.sharesToUnderlying(5 ether);
        assertEq(underlying, 5 ether);

        uint256 underlying1 = theweb3ChainBase.sharesToUnderlyingView(5 ether);
        assertEq(underlying1, 5 ether);

        uint256 share = theweb3ChainBase.underlyingToSharesView(1 ether);
        assertEq(share, 1 ether);

        uint256 share1 = theweb3ChainBase.underlyingToShares(1 ether);
        assertEq(share1, 1 ether);
    }

    function testExplanation() public view {
        string memory expectResult = "theweb3Chain Pos Staking Protocol";
        string memory result = theweb3ChainBase.explanation();
        assertEq(result, expectResult);
    }

    function testPaused() public {
        vm.prank(pauser1);
        theweb3ChainBase.pauseAll();

        vm.prank(strategyManager);
        vm.expectRevert("Pausable: contract is paused");
        theweb3ChainBase.withdraw(user1, 1 ether);
    }
}
