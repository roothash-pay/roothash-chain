// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "@openzeppelin-upgrades/contracts/utils/ReentrancyGuardUpgradeable.sol";

import "./theweb3ChainDepositManagerStorage.sol";
import "../../libraries/EIP1271SignatureUtils.sol";
import "@/access/Pausable.sol";

contract theweb3ChainDepositManager is
    Initializable,
    OwnableUpgradeable,
    ReentrancyGuardUpgradeable,
    Pausable,
    theweb3ChainDepositManagerStorage
{
    uint8 internal constant PAUSED_DEPOSITS = 0;
    uint256 internal ORIGINAL_CHAIN_ID;

    modifier onlyDelegationManager() {
        require(msg.sender == address(delegation), "onlyDelegationManager");
        _;
    }

    constructor() {
        _disableInitializers();
    }

    function initialize(address initialOwner, IDelegationManager _delegation, Itheweb3ChainBase _theweb3ChainBase)
        public
        initializer
    {
        _DOMAIN_SEPARATOR = _calculateDomainSeparator();
        __Ownable_init(initialOwner);
        _inittheweb3ChainDepositManagerStorage(_delegation, _theweb3ChainBase);
    }

    receive() external payable {}

    function depositIntotheweb3Chain(uint256 amount)
        external
        payable
        whenNotPaused
        nonReentrant
        returns (uint256 shares)
    {
        require(amount == msg.value, "deposit value not match amount");
        shares = _depositIntotheweb3Chain(msg.sender, amount);
    }

    function depositIntotheweb3ChainWithSignature(
        uint256 amount,
        address staker,
        uint256 expiry,
        bytes memory signature
    ) external payable whenNotPaused nonReentrant returns (uint256 shares) {
        require(amount == msg.value, "deposit value not match amount");
        require(
            expiry >= block.timestamp,
            "theweb3ChainDepositManager.depositIntotheweb3ChainWithSignature: signature expired"
        );
        uint256 nonce = nonces[staker];

        bytes32 structHash = keccak256(abi.encode(DEPOSIT_TYPEHASH, staker, amount, nonce, expiry));

        unchecked {
            nonces[staker] = nonce + 1;
        }

        bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", domainSeparator(), structHash));

        EIP1271SignatureUtils.checkSignature_EIP1271(staker, digestHash, signature);

        shares = _depositIntotheweb3Chain(staker, amount);
    }

    function removeShares(address staker, uint256 shareAmount) external onlyDelegationManager {
        require(shareAmount != 0, "theweb3ChainDepositManager.removeShares: shareAmount should not be zero!");

        uint256 userShares = stakertheweb3ChainBaseShares[staker];

        require(shareAmount <= userShares, "theweb3ChainDepositManager._removeShares: shareAmount too high");

        unchecked {
            userShares = userShares - shareAmount;
        }

        stakertheweb3ChainBaseShares[staker] = userShares;

        if (userShares == 0) {
            delete stakertheweb3ChainBaseShares[staker];
            theweb3ChainBase.deleteStaker(staker);
        }
    }

    function addShares(address staker, uint256 shares) external onlyDelegationManager {
        _addShares(staker, shares);
    }

    function withdrawSharesAsCp(address recipient, uint256 shares) external onlyDelegationManager {
        theweb3ChainBase.withdraw(recipient, shares);
    }

    function getDeposits(address staker) external view returns (uint256) {
        return stakertheweb3ChainBaseShares[staker];
    }

    function domainSeparator() public view returns (bytes32) {
        if (block.chainid == ORIGINAL_CHAIN_ID) {
            return _DOMAIN_SEPARATOR;
        } else {
            return _calculateDomainSeparator();
        }
    }

    // ================= internal function =================
    function _depositIntotheweb3Chain(address staker, uint256 amount) internal returns (uint256 shares) {
        shares = theweb3ChainBase.deposit{value: amount}(amount, staker);

        _addShares(staker, shares);

        delegation.increaseDelegatedShares(staker, shares);

        return shares;
    }

    function _addShares(address staker, uint256 shares) internal {
        require(staker != address(0), "theweb3ChainDepositManager._addShares: staker cannot be zero address");
        require(shares != 0, "theweb3ChainDepositManager._addShares: shares should not be zero!");
        stakertheweb3ChainBaseShares[staker] += shares;
        emit Deposit(staker, theweb3ChainBase, shares);
    }

    function _calculateDomainSeparator() internal view returns (bytes32) {
        return keccak256(abi.encode(DOMAIN_TYPEHASH, keccak256(bytes("theweb3Chain")), block.chainid, address(this)));
    }
}
