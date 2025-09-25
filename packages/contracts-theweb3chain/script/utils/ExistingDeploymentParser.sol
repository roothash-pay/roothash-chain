// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import "../../src/core/pos/theweb3ChainBase.sol";
import "../../src/core/pos/DelegationManager.sol";
import "../../src/core/pos/RewardManager.sol";
import "../../src/core/pos/theweb3ChainBase.sol";
import "../../src/core/pos/theweb3ChainDepositManager.sol";
import "../../src/core/pos/SlashingManager.sol";

import "../../src/access/PauserRegistry.sol";

import "../utils/EmptyContract.sol";

import "forge-std/Script.sol";
import "forge-std/Test.sol";

struct CpTokenConfig {
    address tokenAddress;
    string tokenName;
    string tokenSymbol;
}

contract ExistingDeploymentParser is Script, Test {
    // theweb3ChainLayer Contracts
    ProxyAdmin public theweb3ChainLayerProxyAdmin;
    PauserRegistry public theweb3ChainLayerPauserReg;

    DelegationManager public delegationManager;
    ProxyAdmin public delegationManagerProxyAdmin;
    DelegationManager public delegationManagerImplementation;

    theweb3ChainDepositManager public theweb3ChainDepositManager;
    ProxyAdmin public theweb3ChainDepositManagerProxyAdmin;
    theweb3ChainDepositManager public theweb3ChainDepositManagerImplementation;

    RewardManager public rewardManager;
    ProxyAdmin public rewardManagerProxyAdmin;
    RewardManager public rewardManagerImplementation;

    SlashingManager public slashingManager;
    ProxyAdmin public slashingManagerProxyAdmin;
    SlashingManager public slashingManagerImplementation;

    theweb3ChainBase public theweb3ChainBase;
    ProxyAdmin public chainBaseBaseProxyAdmin;
    theweb3ChainBase public theweb3ChainBaseImplementation;

    EmptyContract public emptyContract;

    // Reward Token
    address public rewardTokenAddress;

    address executorMultisig;
    address operationsMultisig;
    address communityMultisig;
    address pauserMultisig;
    address timelock;

    /// @notice Initialization Params for first initial deployment scripts
    // theweb3ChainDepositManager
    uint256 theweb3ChainDEPOSITMANAGER_MANAGER_INIT_PAUSED_STATUS;

    // DelegationManager
    uint256 DELEGATION_MANAGER_INIT_PAUSED_STATUS;
    uint256 DELEGATION_MANAGER_WITHDRAWAL_DELAY_BLOCK;

    // RewardManager
    uint256 REWARD_MANAGER_INIT_PAUSED_STATUS;
    uint32 REWARD_MANAGER_MAX_REWARDS_DURATION;
    uint32 REWARD_MANAGER_MAX_RETROACTIVE_LENGTH;
    uint32 REWARD_MANAGER_MAX_FUTURE_LENGTH;
    uint32 REWARD_MANAGER_GENESIS_REWARDS_TIMESTAMP;
    address REWARD_MANAGER_UPDATER;
    uint32 REWARD_MANAGER_ACTIVATION_DELAY;
    uint32 REWARD_MANAGER_CALCULATION_INTERVAL_SECONDS;
    uint32 REWARD_MANAGER_GLOBAL_OPERATOR_COMMISSION_BIPS;
    address REWARD_MANAGER_RWARD_TOKEN_ADDRESS;
    uint32 REWARD_MANAGER_STAKE_PERCENTAGE;

    // one week in blocks -- 50400
    uint32 DELAYED_WITHDRAWAL_ROUTER_INIT_WITHDRAWAL_DELAY_BLOCKS;

    // theweb3ChainBase Deployment
    uint256 theweb3ChainBASE_MIN_DEPOSIT;
    uint256 theweb3ChainBASE_MAX_DEPOSIT;

    /// @notice use for parsing already deployed theweb3ChainLayer contracts
    function _parseDeployedContracts(
        string memory existingDeploymentInfoPath
    ) internal virtual {
        // read and log the chainID
        uint256 currentChainId = block.chainid;
        emit log_named_uint("You are parsing on ChainID", currentChainId);

        // READ JSON CONFIG DATA
        string memory existingDeploymentData = vm.readFile(
            existingDeploymentInfoPath
        );

        // check that the chainID matches the one in the config
        uint256 configChainId = stdJson.readUint(
            existingDeploymentData,
            ".chainInfo.chainId"
        );
        require(
            configChainId == currentChainId,
            "You are on the wrong chain for this config"
        );

        // read all of the deployed addresses
        executorMultisig = stdJson.readAddress(
            existingDeploymentData,
            ".parameters.executorMultisig"
        );
        operationsMultisig = stdJson.readAddress(
            existingDeploymentData,
            ".parameters.operationsMultisig"
        );
        communityMultisig = stdJson.readAddress(
            existingDeploymentData,
            ".parameters.communityMultisig"
        );
        pauserMultisig = stdJson.readAddress(
            existingDeploymentData,
            ".parameters.pauserMultisig"
        );
        timelock = stdJson.readAddress(
            existingDeploymentData,
            ".parameters.timelock"
        );

        theweb3ChainLayerProxyAdmin = ProxyAdmin(
            stdJson.readAddress(
                existingDeploymentData,
                ".addresses.theweb3ChainLayerProxyAdmin"
            )
        );
        theweb3ChainLayerPauserReg = PauserRegistry(
            stdJson.readAddress(
                existingDeploymentData,
                ".addresses.theweb3ChainLayerPauserReg"
            )
        );

        delegationManager = DelegationManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.delegationManager"
                )
            )
        );
        delegationManagerImplementation = DelegationManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.delegationManagerImplementation"
                )
            )
        );

        rewardManager = RewardManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.rewardManager"
                )
            )
        );
        rewardManagerImplementation = RewardManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.rewardManagerImplementation"
                )
            )
        );

        theweb3ChainDepositManager = theweb3ChainDepositManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.theweb3ChainDepositManager"
                )
            )
        );
        theweb3ChainDepositManagerImplementation = theweb3ChainDepositManager(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.theweb3ChainDepositManagerImplementation"
                )
            )
        );

        theweb3ChainBase = theweb3ChainBase(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.theweb3ChainBase"
                )
            )
        );
        theweb3ChainBaseImplementation = theweb3ChainBase(
            payable(
                stdJson.readAddress(
                    existingDeploymentData,
                    ".addresses.theweb3ChainBaseImplementation"
                )
            )
        );

        emptyContract = EmptyContract(
            stdJson.readAddress(
                existingDeploymentData,
                ".addresses.emptyContract"
            )
        );
    }

    /// @notice use for deploying a new set of theweb3ChainLayer contracts
    /// Note that this does require multisigs to already be deployed
    function _parseInitialDeploymentParams(
        string memory initialDeploymentParamsPath
    ) internal virtual {
        // read and log the chainID
        uint256 currentChainId = block.chainid;
        emit log_named_uint("You are parsing on ChainID", currentChainId);

        // READ JSON CONFIG DATA
        string memory initialDeploymentData = vm.readFile(
            initialDeploymentParamsPath
        );

        // check that the chainID matches the one in the config
        uint256 configChainId = stdJson.readUint(
            initialDeploymentData,
            ".chainInfo.chainId"
        );
        require(
            configChainId == currentChainId,
            "You are on the wrong chain for this config"
        );

        // read all of the deployed addresses
        executorMultisig = stdJson.readAddress(
            initialDeploymentData,
            ".multisig_addresses.executorMultisig"
        );
        operationsMultisig = stdJson.readAddress(
            initialDeploymentData,
            ".multisig_addresses.operationsMultisig"
        );
        communityMultisig = stdJson.readAddress(
            initialDeploymentData,
            ".multisig_addresses.communityMultisig"
        );
        pauserMultisig = stdJson.readAddress(
            initialDeploymentData,
            ".multisig_addresses.pauserMultisig"
        );

        // ChainBases to Deploy, load chainBase list
        theweb3ChainBASE_MIN_DEPOSIT = stdJson.readUint(
            initialDeploymentData,
            ".theweb3ChainBase.MIN_DEPOSIT"
        );
        theweb3ChainBASE_MAX_DEPOSIT = stdJson.readUint(
            initialDeploymentData,
            ".theweb3ChainBase.MAX_DEPOSIT"
        );

        // Read initialize params for upgradeable contracts
        theweb3ChainDEPOSITMANAGER_MANAGER_INIT_PAUSED_STATUS = stdJson.readUint(
            initialDeploymentData,
            ".theweb3ChainDepositManager.init_paused_status"
        );

        // DelegationManager
        DELEGATION_MANAGER_WITHDRAWAL_DELAY_BLOCK = stdJson.readUint(
            initialDeploymentData,
            ".delegationManager.withdrawalDelayBlock"
        );
        DELEGATION_MANAGER_INIT_PAUSED_STATUS = stdJson.readUint(
            initialDeploymentData,
            ".delegationManager.init_paused_status"
        );
        // RewardManager
        REWARD_MANAGER_INIT_PAUSED_STATUS = stdJson.readUint(
            initialDeploymentData,
            ".rewardManager.init_paused_status"
        );
        REWARD_MANAGER_CALCULATION_INTERVAL_SECONDS = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.CALCULATION_INTERVAL_SECONDS"
            )
        );
        REWARD_MANAGER_MAX_REWARDS_DURATION = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.MAX_REWARDS_DURATION"
            )
        );
        REWARD_MANAGER_MAX_RETROACTIVE_LENGTH = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.MAX_RETROACTIVE_LENGTH"
            )
        );
        REWARD_MANAGER_MAX_FUTURE_LENGTH = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.MAX_FUTURE_LENGTH"
            )
        );
        REWARD_MANAGER_GENESIS_REWARDS_TIMESTAMP = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.GENESIS_REWARDS_TIMESTAMP"
            )
        );
        REWARD_MANAGER_UPDATER = stdJson.readAddress(
            initialDeploymentData,
            ".rewardManager.rewards_updater_address"
        );
        REWARD_MANAGER_ACTIVATION_DELAY = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.activation_delay"
            )
        );
        REWARD_MANAGER_GLOBAL_OPERATOR_COMMISSION_BIPS = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.global_operator_commission_bips"
            )
        );
        REWARD_MANAGER_STAKE_PERCENTAGE = uint32(
            stdJson.readUint(
                initialDeploymentData,
                ".rewardManager.stake_percentage"
            )
        );

        logInitialDeploymentParams();
    }

    /// @notice Ensure contracts point at each other correctly via constructors
    function _verifyContractPointers() internal view virtual {
        // RewardManager
        require(
            rewardManager.delegationManager() == delegationManager,
            "rewardManager: delegationManager address not set correctly"
        );
        require(
            rewardManager.theweb3ChainDepositManager() == theweb3ChainDepositManager,
            "rewardManager: theweb3ChainDepositManager address not set correctly"
        );
        // DelegationManager
        require(
            delegationManager.theweb3ChainDepositManager() == theweb3ChainDepositManager,
            "delegationManager: theweb3ChainDepositManager address not set correctly"
        );
        // theweb3ChainDepositManager
        require(
            theweb3ChainDepositManager.delegation() == delegationManager,
            "theweb3ChainDepositManager: delegationManager address not set correctly"
        );
    }

    /// @notice verify implementations for Transparent Upgradeable Proxies
    /// Note that the instance of ProxyAdmin can no longer invoke {getProxyImplementation} in the dependencies from the latest version of OpenZeppelin
    function _verifyImplementations() internal view virtual {
        require(
            getImplementationAddress(address(rewardManager)) ==
                address(rewardManagerImplementation),
            "rewardManager: implementation set incorrectly"
        );
        require(
            getImplementationAddress(address(delegationManager)) ==
                address(delegationManagerImplementation),
            "delegationManager: implementation set incorrectly"
        );
        require(
            getImplementationAddress(address(theweb3ChainDepositManager)) ==
                address(theweb3ChainDepositManagerImplementation),
            "theweb3ChainDepositManager: implementation set incorrectly"
        );
    }

    /**
     * @notice Verify initialization of Transparent Upgradeable Proxies. Also check
     * initialization params if this is the first deployment.
     */
    function _verifyContractsInitialized() internal virtual {
        // RewardManager
        vm.expectRevert(abi.encodeWithSignature("InvalidInitialization()"));
        rewardManager.initialize(
            executorMultisig,
            executorMultisig,
            executorMultisig,
            REWARD_MANAGER_STAKE_PERCENTAGE,
            theweb3ChainLayerPauserReg,
            delegationManager,
            theweb3ChainDepositManager
        );

        // DelegationManager
        vm.expectRevert(abi.encodeWithSignature("InvalidInitialization()"));
        delegationManager.initialize(
            address(0),
            theweb3ChainLayerPauserReg,
            0,
            0,
            theweb3ChainDepositManager,
            theweb3ChainBase,
            slashingManager
        );

        // theweb3ChainDepositManager
        vm.expectRevert(abi.encodeWithSignature("InvalidInitialization()"));
        theweb3ChainDepositManager.initialize(
            address(0),
            delegationManager,
            theweb3ChainBase
        );

        // ChainBases
        vm.expectRevert(abi.encodeWithSignature("InvalidInitialization()"));
        theweb3ChainBase(address(theweb3ChainBase)).initialize(
            theweb3ChainLayerPauserReg,
            0,
            0,
            theweb3ChainDepositManager
        );
    }

    /// @notice Verify params based on config constants that are updated from calling `_parseInitialDeploymentParams`
    function _verifyInitializationParams() internal view virtual {
        // RewardManager
        require(
            rewardManager.owner() == executorMultisig,
            "rewardManager: owner not set correctly"
        );
        // DelegationManager
        require(
            delegationManager.pauserRegistry() == theweb3ChainLayerPauserReg,
            "delegationManager: pauser registry not set correctly"
        );
        require(
            delegationManager.owner() == executorMultisig,
            "delegationManager: owner not set correctly"
        );
        require(
            delegationManager.paused() == DELEGATION_MANAGER_INIT_PAUSED_STATUS,
            "delegationManager: init paused status set incorrectly"
        );

        // theweb3ChainDepositManager
        require(
            theweb3ChainDepositManager.owner() == executorMultisig,
            "theweb3ChainDepositManager: owner not set correctly"
        );

        // ChainBases
        require(
            theweb3ChainBase.pauserRegistry() == theweb3ChainLayerPauserReg,
            "theweb3ChainBase: pauser registry not set correctly"
        );
        require(
            theweb3ChainBase.paused() == 0,
            "theweb3ChainBase: init paused status set incorrectly"
        );

        // Pausing Permissions
        require(
            theweb3ChainLayerPauserReg.isPauser(operationsMultisig),
            "pauserRegistry: operationsMultisig is not pauser"
        );
        require(
            theweb3ChainLayerPauserReg.isPauser(executorMultisig),
            "pauserRegistry: executorMultisig is not pauser"
        );
        require(
            theweb3ChainLayerPauserReg.isPauser(pauserMultisig),
            "pauserRegistry: pauserMultisig is not pauser"
        );
        require(
            theweb3ChainLayerPauserReg.unpauser() == executorMultisig,
            "pauserRegistry: unpauser not set correctly"
        );
    }

    function logInitialDeploymentParams() public {
        emit log_string(
            "==== Parsed Initilize Params for Initial Deployment ===="
        );

        emit log_named_address("executorMultisig", executorMultisig);
        emit log_named_address("operationsMultisig", operationsMultisig);
        emit log_named_address("communityMultisig", communityMultisig);
        emit log_named_address("pauserMultisig", pauserMultisig);

        emit log_named_uint(
            "theweb3ChainDEPOSITMANAGER_MANAGER_INIT_PAUSED_STATUS",
            theweb3ChainDEPOSITMANAGER_MANAGER_INIT_PAUSED_STATUS
        );
        emit log_named_uint(
            "DELEGATION_MANAGER_WITHDRAWAL_DELAY_BLOCK",
            DELEGATION_MANAGER_WITHDRAWAL_DELAY_BLOCK
        );
        emit log_named_uint(
            "DELEGATION_MANAGER_INIT_PAUSED_STATUS",
            DELEGATION_MANAGER_INIT_PAUSED_STATUS
        );
        emit log_named_uint(
            "REWARD_MANAGER_INIT_PAUSED_STATUS",
            REWARD_MANAGER_INIT_PAUSED_STATUS
        );
        emit log_named_uint(
            "DELAYED_WITHDRAWAL_ROUTER_INIT_WITHDRAWAL_DELAY_BLOCKS",
            DELAYED_WITHDRAWAL_ROUTER_INIT_WITHDRAWAL_DELAY_BLOCKS
        );
    }

    /**
     * @notice Log contract addresses and write to output json file
     */
    function logAndOutputContractAddresses(string memory outputPath) public {
        string memory parent_object = "parent object";

        string memory deployed_addresses = "addresses";
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainLayerProxyAdmin",
            address(theweb3ChainLayerProxyAdmin)
        );
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainLayerPauserReg",
            address(theweb3ChainLayerPauserReg)
        );
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainBase",
            address(theweb3ChainBase)
        );
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainBaseImplementation",
            address(theweb3ChainBaseImplementation)
        );
        vm.serializeAddress(
            deployed_addresses,
            "delegationManager",
            address(delegationManager)
        );
        vm.serializeAddress(
            deployed_addresses,
            "delegationManagerImplementation",
            address(delegationManagerImplementation)
        );
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainDepositManager",
            address(theweb3ChainDepositManager)
        );
        vm.serializeAddress(
            deployed_addresses,
            "theweb3ChainDepositManagerImplementation",
            address(theweb3ChainDepositManagerImplementation)
        );
        vm.serializeAddress(
            deployed_addresses,
            "rewardManager",
            address(rewardManager)
        );
        vm.serializeAddress(
            deployed_addresses,
            "rewardManagerImplementation",
            address(rewardManagerImplementation)
        );
        vm.serializeAddress(
            deployed_addresses,
            "slashingManager",
            address(slashingManager)
        );
        vm.serializeAddress(
            deployed_addresses,
            "slashingManagerImplementation",
            address(slashingManagerImplementation)
        );
        vm.serializeAddress(
            deployed_addresses,
            "emptyContract",
            address(emptyContract)
        );
        string memory deployed_addresses_output = vm.serializeAddress(
            deployed_addresses,
            "emptyContract",
            address(emptyContract)
        );

        string memory parameters = "parameters";
        vm.serializeAddress(parameters, "executorMultisig", executorMultisig);
        vm.serializeAddress(
            parameters,
            "operationsMultisig",
            operationsMultisig
        );
        vm.serializeAddress(parameters, "communityMultisig", communityMultisig);
        vm.serializeAddress(parameters, "pauserMultisig", pauserMultisig);
        vm.serializeAddress(parameters, "timelock", timelock);
        string memory parameters_output = vm.serializeAddress(
            parameters,
            "operationsMultisig",
            operationsMultisig
        );

        string memory chain_info = "chainInfo";
        vm.serializeUint(chain_info, "deploymentBlock", block.number);
        string memory chain_info_output = vm.serializeUint(
            chain_info,
            "chainId",
            block.chainid
        );

        vm.serializeString(
            parent_object,
            deployed_addresses,
            deployed_addresses_output
        );
        vm.serializeString(parent_object, chain_info, chain_info_output);
        string memory finalJson = vm.serializeString(
            parent_object,
            parameters,
            parameters_output
        );

        vm.writeJson(finalJson, outputPath);
    }

    function getProxyAdminAddress(
        address proxy
    ) internal view returns (address) {
        address CHEATCODE_ADDRESS = 0x7109709ECfa91a80626fF3989D68f67F5b1DD12D;
        Vm vm = Vm(CHEATCODE_ADDRESS);
        bytes32 adminSlot = vm.load(proxy, ERC1967Utils.ADMIN_SLOT);
        return address(uint160(uint256(adminSlot)));
    }

    function getImplementationAddress(
        address proxy
    ) internal view returns (address) {
        address CHEATCODE_ADDRESS = 0x7109709ECfa91a80626fF3989D68f67F5b1DD12D;
        Vm vm = Vm(CHEATCODE_ADDRESS);
        bytes32 implementationSlot = vm.load(
            proxy,
            ERC1967Utils.IMPLEMENTATION_SLOT
        );
        return address(uint160(uint256(implementationSlot)));
    }
}
