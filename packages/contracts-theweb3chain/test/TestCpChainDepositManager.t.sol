// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@/core/pos/theweb3ChainBase.sol";
import "@/core/pos/theweb3ChainDepositManager.sol";
import "@/core/pos/DelegationManager.sol";

import "@/access/PauserRegistry.sol";
import "@/interfaces/IDelegationManager.sol";
import "@/interfaces/ISlashingManager.sol";

import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract theweb3ChainDepositManagerTest is Test {
    theweb3ChainBase public theweb3ChainBase;
    PauserRegistry public pauserregistry;
    theweb3ChainDepositManager public theweb3ChainDepositManager;
    DelegationManager public delegationManager;

    address public user1 = address(0x01);
    address public pauser1 = address(0x02);
    address public pauser2 = address(0x03);
    address[] public pausers;
    address public unpauser = address(0x04);

    address public owner = address(0x05);
    address public slashingManager = address(0x06);

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

        theweb3ChainBase = theweb3ChainBase(payable(address(proxy1)));
        theweb3ChainDepositManager = theweb3ChainDepositManager(payable(address(proxy2)));
        delegationManager = DelegationManager(payable(address(proxy3)));

        theweb3ChainBase.initialize(
            IPauserRegistry(address(pauserregistry)),
            1 ether,
            10 ether,
            Itheweb3ChainDepositManager(address(theweb3ChainDepositManager))
        );
        theweb3ChainDepositManager.initialize(owner, IDelegationManager(address(delegationManager)), theweb3ChainBase);
        delegationManager.initialize(
            owner,
            IPauserRegistry(address(pauserregistry)),
            0,
            0,
            theweb3ChainDepositManager,
            theweb3ChainBase,
            ISlashingManager(address(slashingManager))
        );

        vm.deal(user1, 100 ether);
    }

    function testDepositIntotheweb3Chain() public {
        uint256 beforeShares = theweb3ChainDepositManager.getDeposits(user1);

        vm.startPrank(user1);
        theweb3ChainDepositManager.depositIntotheweb3Chain{value: 2 ether}(2 ether);
        uint256 afterShares = theweb3ChainDepositManager.getDeposits(user1);
        uint256 shares = theweb3ChainBase.shares(user1);
        uint256 value1 = theweb3ChainBase.userUnderlying(user1);
        uint256 value2 = theweb3ChainBase.userUnderlyingView(user1);

        assertGt(afterShares, beforeShares);
        assert(afterShares == 2 ether);
        assert(shares == 2 ether);
        assert(value1 == 2 ether);
        assert(value2 == 2 ether);
        assert(theweb3ChainDepositManager.stakertheweb3ChainBaseShares(user1) == 2 ether);
        vm.stopPrank();
    }

    function testRemoveSharesOnlyDelegationManager() public {
        vm.prank(user1);
        theweb3ChainDepositManager.depositIntotheweb3Chain{value: 2 ether}(2 ether);

        vm.prank(address(delegationManager));
        theweb3ChainDepositManager.removeShares(user1, 1 ether);
        assertEq(theweb3ChainDepositManager.getDeposits(user1), 1 ether);

        vm.expectRevert("onlyDelegationManager");
        theweb3ChainDepositManager.removeShares(user1, 1 ether);

        vm.expectRevert("theweb3ChainDepositManager.removeShares: shareAmount should not be zero!");
        vm.prank(address(delegationManager));
        theweb3ChainDepositManager.removeShares(user1, 0);

        vm.expectRevert("theweb3ChainDepositManager._removeShares: shareAmount too high");
        vm.prank(address(delegationManager));
        theweb3ChainDepositManager.removeShares(user1, 100 ether);
    }

    function testWithdrawSharesAsCpOnlyDelegationManager() public {
        vm.startPrank(user1);
        theweb3ChainDepositManager.depositIntotheweb3Chain{value: 2 ether}(2 ether);
        vm.stopPrank();

        vm.prank(address(delegationManager));
        theweb3ChainDepositManager.withdrawSharesAsCp(user1, 1 ether);

        assertEq(user1.balance, 99 ether); // 或 assertGt(user.balance, 0);

        vm.expectRevert("onlyDelegationManager");
        theweb3ChainDepositManager.withdrawSharesAsCp(user1, 1 ether);
    }

    function testAddShares() public {
        vm.prank(address(delegationManager));
        theweb3ChainDepositManager.addShares(user1, 5 ether);

        uint256 stored = theweb3ChainDepositManager.getDeposits(user1);
        assertEq(stored, 5 ether);

        vm.prank(address(delegationManager));
        vm.expectRevert("theweb3ChainDepositManager._addShares: shares should not be zero!");
        theweb3ChainDepositManager.addShares(user1, 0);

        vm.expectRevert("onlyDelegationManager");
        theweb3ChainDepositManager.addShares(user1, 1 ether);
    }

    function testSignatureWithVmSign() public pure {
        uint256 privKey = 0x4f3edf983ac636a65a842ce7c78d9aa706d3b113b37c9430e6fd8c8d11f8b4e6;
        address expectedSigner = vm.addr(privKey);

        // 构造 EIP-712 的 digestHash（你可以通过 helper 函数生成）
        bytes32 digest = 0x349dbac25db23f8be0b94a49883a681b0deebbb21ec6eb1c4a2676c9d5e1e222;

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privKey, digest);

        bytes memory signature = abi.encodePacked(r, s, v);

        address result = ECDSA.recover(digest, signature);
        assertEq(result, expectedSigner);
    }

    function testdepositIntotheweb3ChainWithSignature() public {
        uint256 privKey = 0x4f3edf983ac636a65a842ce7c78d9aa706d3b113b37c9430e6fd8c8d11f8b4e6;
        address staker = vm.addr(privKey);

        bytes32 DEPOSIT_TYPEHASH =
            keccak256("Deposit(address staker,address theweb3ChainBase,uint256 amount,uint256 nonce,uint256 expiry)");
        bytes32 DOMAIN_TYPEHASH = keccak256("EIP712Domain(string name,uint256 chainId,address verifyingContract)");

        bytes32 structHash = keccak256(abi.encode(DEPOSIT_TYPEHASH, staker, 2 ether, 0, 100));

        bytes32 domainSeparator = keccak256(
            abi.encode(DOMAIN_TYPEHASH, keccak256("theweb3Chain"), block.chainid, address(theweb3ChainDepositManager))
        );
        bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", domainSeparator, structHash));
        bytes32 digest = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", digestHash));

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privKey, digest);
        bytes memory signature = abi.encodePacked(r, s, v);

        vm.prank(user1);
        theweb3ChainDepositManager.depositIntotheweb3ChainWithSignature{value: 2 ether}(2 ether, staker, 100, signature);

        uint256 afterShares = theweb3ChainDepositManager.getDeposits(staker);
        assert(afterShares == 2 ether);

        vm.prank(user1);
        vm.expectRevert("deposit value not match amount");
        theweb3ChainDepositManager.depositIntotheweb3ChainWithSignature{value: 2 ether}(
            10 ether, staker, 100, signature
        );

        vm.prank(user1);
        vm.warp(10000);
        vm.expectRevert("theweb3ChainDepositManager.depositIntotheweb3ChainWithSignature: signature expired");
        theweb3ChainDepositManager.depositIntotheweb3ChainWithSignature{value: 2 ether}(2 ether, staker, 100, signature);
    }
}
