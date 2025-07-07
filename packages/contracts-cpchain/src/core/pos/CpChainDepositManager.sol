// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "@openzeppelin-upgrades/contracts/utils/ReentrancyGuardUpgradeable.sol";

import "./CpChainDepositManagerStorage.sol";
import "../../libraries/EIP1271SignatureUtils.sol";

contract CpChainDepositManager is Initializable, OwnableUpgradeable, ReentrancyGuardUpgradeable, CpChainDepositManagerStorage {
    uint8 internal constant PAUSED_DEPOSITS = 0;
    uint256 internal immutable ORIGINAL_CHAIN_ID;

    modifier onlyDelegationManager() {
        require(msg.sender == address(delegation), "onlyDelegationManager");
        _;
    }

    modifier onlyCpChainBaseWhitelister() {
        require(
            msg.sender == cpChainWhitelister,
            "CpChainDepositManager.onlyCpChainBaseWhitelister: not the cpChainWhitelister"
        );
        _;
    }

    modifier onlyStrategiesWhitelistedForDeposit(ICpChainBase chainBase) {
        require(
            cpChainWhitelistedForDeposit[chainBase],
            "CpChainDepositManager.onlyStrategiesWhitelistedForDeposit: chainBase not whitelisted"
        );
        _;
    }

    constructor(IDelegationManager _delegation) CpChainDepositManagerStorage(_delegation) {
        _disableInitializers();
    }

    // EXTERNAL FUNCTIONS
    function initialize(
        address initialOwner,
        address initialCpChainBaseWhitelister
    ) external initializer {
        _DOMAIN_SEPARATOR = _calculateDomainSeparator();
        _transferOwnership(initialOwner);
        _setCpChainBaseWhitelister(initialCpChainBaseWhitelister);
    }

    function depositIntoCpChain(
        ICpChainBase chainBase,
        uint256 amount
    ) external payable nonReentrant returns (uint256 shares)  {
        shares = _depositIntoCpChain(msg.sender, chainBase, amount);
    }

    function depositIntoCpChainWithSignature(
        ICpChainBase chainBase,
        uint256 amount,
        address staker,
        uint256 expiry,
        bytes memory signature
    ) external payable nonReentrant returns (uint256 shares) {
        require(
            !thirdPartyTransfersForbidden[chainBase],
            "CpChainDepositManager.depositIntoCpChainWithSignature: third transfers disabled"
        );
        require(expiry >= block.timestamp, "CpChainDepositManager.depositIntoCpChainWithSignature: signature expired");
        uint256 nonce = nonces[staker];
        bytes32 structHash = keccak256(abi.encode(DEPOSIT_TYPEHASH, staker, chainBase, amount, nonce, expiry));
        unchecked {
            nonces[staker] = nonce + 1;
        }

        bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", domainSeparator(), structHash));

        EIP1271SignatureUtils.checkSignature_EIP1271(staker, digestHash, signature);

        shares = _depositIntoCpChain(staker, chainBase, amount);
    }

    function removeShares(
        address staker,
        ICpChainBase chainBase,
        uint256 shares
    ) external onlyDelegationManager {
        _removeShares(staker, chainBase, shares);
    }

    function addShares(
        address staker,
        ICpChainBase chainBase,
        uint256 shares
    ) external onlyDelegationManager {
        _addShares(staker, chainBase, shares);
    }

    function withdrawSharesAsCp(
        address recipient,
        ICpChainBase chainBase,
        uint256 shares
    ) external onlyDelegationManager {
        chainBase.withdraw(recipient, shares);
    }

    function getStakerCpChainBaseShares(address user, ICpChainBase chainBase) external view returns (uint256 shares) {
        return stakerCpChainBaseShares[user][chainBase];
    }

    function setCpChainBaseWhitelister(address newCpChainBaseWhitelister) external onlyOwner {
        _setCpChainBaseWhitelister(newCpChainBaseWhitelister);
    }

    function addStrategiesToDepositWhitelist(
        ICpChainBase[] calldata strategiesToWhitelist,
        bool[] calldata thirdPartyTransfersForbiddenValues
    ) external onlyCpChainBaseWhitelister {
        require(
            strategiesToWhitelist.length == thirdPartyTransfersForbiddenValues.length,
            "CpChainDepositManager.addStrategiesToDepositWhitelist: array lengths do not match"
        );
        uint256 strategiesToWhitelistLength = strategiesToWhitelist.length;
        for (uint256 i = 0; i < strategiesToWhitelistLength; ) {
            if (!cpChainWhitelistedForDeposit[strategiesToWhitelist[i]]) {
                cpChainWhitelistedForDeposit[strategiesToWhitelist[i]] = true;
                emit CpChainBaseAddedToDepositWhitelist(strategiesToWhitelist[i]);
                _setThirdPartyTransfersForbidden(strategiesToWhitelist[i], thirdPartyTransfersForbiddenValues[i]);
            }
            unchecked {
                ++i;
            }
        }
    }

    function removeStrategiesFromDepositWhitelist(
        ICpChainBase[] calldata strategiesToRemoveFromWhitelist
    ) external onlyCpChainBaseWhitelister {
        uint256 strategiesToRemoveFromWhitelistLength = strategiesToRemoveFromWhitelist.length;
        for (uint256 i = 0; i < strategiesToRemoveFromWhitelistLength; ) {
            if (cpChainWhitelistedForDeposit[strategiesToRemoveFromWhitelist[i]]) {
                cpChainWhitelistedForDeposit[strategiesToRemoveFromWhitelist[i]] = false;
                emit CpChainBaseRemovedFromDepositWhitelist(strategiesToRemoveFromWhitelist[i]);
                _setThirdPartyTransfersForbidden(strategiesToRemoveFromWhitelist[i], false);
            }
            unchecked {
                ++i;
            }
        }
    }

    // INTERNAL FUNCTIONS
    function _addShares(address staker, ICpChainBase chainBase, uint256 shares) internal {
        require(staker != address(0), "CpChainDepositManager._addShares: staker cannot be zero address");
        require(shares != 0, "CpChainDepositManager._addShares: shares should not be zero!");

        if (stakerCpChainBaseShares[staker][chainBase] == 0) {
            require(
                stakerCpChainBaseList[staker].length < MAX_STAKER_STRATEGY_LIST_LENGTH,
                "CpChainDepositManager._addShares: deposit would exceed MAX_STAKER_STRATEGY_LIST_LENGTH"
            );
            stakerCpChainBaseList[staker].push(chainBase);
        }

        stakerCpChainBaseShares[staker][chainBase] += shares;

        emit Deposit(staker, chainBase, shares);
    }

    function _depositIntoCpChain(
        address staker,
        ICpChainBase chainBase,
        uint256 amount
    ) internal onlyStrategiesWhitelistedForDeposit(chainBase) returns (uint256 shares) {
        shares = chainBase.deposit(amount);

        _addShares(staker, chainBase, shares);

        delegation.increaseDelegatedShares(staker, chainBase, shares);

        return shares;
    }

    function _removeShares(
        address staker,
        ICpChainBase chainBase,
        uint256 shareAmount
    ) internal returns (bool) {
        require(shareAmount != 0, "CpChainDepositManager._removeShares: shareAmount should not be zero!");

        uint256 userShares = stakerCpChainBaseShares[staker][chainBase];

        require(shareAmount <= userShares, "CpChainDepositManager._removeShares: shareAmount too high");
        unchecked {
            userShares = userShares - shareAmount;
        }

        stakerCpChainBaseShares[staker][chainBase] = userShares;

        if (userShares == 0) {
            _removeCpChainBaseFromStakerCpChainBaseList(staker, chainBase);
            return true;
        }
        return false;
    }

    function _removeCpChainBaseFromStakerCpChainBaseList(
        address staker,
        ICpChainBase chainBase
    ) internal {
        uint256 cpChainsLength = stakerCpChainBaseList[staker].length;
        uint256 j = 0;
        for (; j < cpChainsLength; ) {
            if (stakerCpChainBaseList[staker][j] == chainBase) {
                stakerCpChainBaseList[staker][j] = stakerCpChainBaseList[staker][stakerCpChainBaseList[staker].length - 1];
                break;
            }
            unchecked { ++j; }
        }
        require(j != cpChainsLength, "CpChainDepositManager._removeCpChainBaseFromStakerCpChainBaseList: chainBase not found");
        stakerCpChainBaseList[staker].pop();
    }

    function _setThirdPartyTransfersForbidden(ICpChainBase chainBase, bool value) internal {
        emit UpdatedThirdPartyTransfersForbidden(chainBase, value);
        thirdPartyTransfersForbidden[chainBase] = value;
    }

    function _setCpChainBaseWhitelister(address newCpChainBaseWhitelister) internal {
        emit CpChainBaseWhitelisterChanged(cpChainWhitelister, newCpChainBaseWhitelister);
        cpChainWhitelister = newCpChainBaseWhitelister;
    }

    // VIEW FUNCTIONS
    function getDeposits(address staker) external view returns (ICpChainBase[] memory, uint256[] memory) {
        uint256 strategiesLength = stakerCpChainBaseList[staker].length;
        uint256[] memory shares = new uint256[](strategiesLength);
        for (uint256 i = 0; i < strategiesLength; ) {
            shares[i] = stakerCpChainBaseShares[staker][stakerCpChainBaseList[staker][i]];
            unchecked {
                ++i;
            }
        }
        return (stakerCpChainBaseList[staker], shares);
    }

    function stakerCpChainBaseListLength(address staker) external view returns (uint256) {
        return stakerCpChainBaseList[staker].length;
    }

    function domainSeparator() public view returns (bytes32) {
        if (block.chainid == ORIGINAL_CHAIN_ID) {
            return _DOMAIN_SEPARATOR;
        } else {
            return _calculateDomainSeparator();
        }
    }

    function _calculateDomainSeparator() internal view returns (bytes32) {
        return keccak256(abi.encode(DOMAIN_TYPEHASH, keccak256(bytes("CpChain")), block.chainid, address(this)));
    }
}
