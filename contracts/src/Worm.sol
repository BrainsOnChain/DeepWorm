// SPDX-License-Identifier: MIT

pragma solidity ^0.8.28;

import {IMarlinTEEAttestationVerifier} from "./IMarlinTEEAttestationVerifier.sol";

contract Worm {
    IMarlinTEEAttestationVerifier public immutable ATTESTATION_VERIFIER;
    uint256 public immutable UPDATE_COOLDOWN_TIME;
    uint256 public immutable TRIGGER_COOLDOWN_TIME;

    bytes public pcrs;

    uint256 public lastUpdatedTimestamp;
    uint256 public lastTriggeredTimestamp;
    address public enclave;
    WormState public wormState;

    struct WormState {
        uint256 leftMuscle;
        uint256 rightMuscle;
    }

    event WormStateUpdated(
        int256 deltaX,
        int256 deltaY,
        uint256 leftMuscle,
        uint256 rightMuscle,
        uint256 positionTimestamp,
        uint256 positionPrice
    );
    event WormStateUpdatedByUser(
        int256 deltaX,
        int256 deltaY,
        uint256 leftMuscle,
        uint256 rightMuscle,
        uint256 positionTimestamp,
        address indexed triggeringUser
    );
    event EnclaveKeyUpdated(address indexed enclave);
    event UserTriggeredWorm(address indexed triggeringUser);

    error PubkeyLengthInvalid();
    error UpdateCooldownNotOver();
    error TriggerCooldownNotOver();
    error InvalidCaller();
    error InvalidPcrs();

    constructor(
        address _attestationVerifier,
        uint256 _updateCooldownTime,
        uint256 _triggerCooldownTime,
        bytes memory _pcrs
    ) payable {
        require(_pcrs.length == 144, InvalidPcrs());
        ATTESTATION_VERIFIER = IMarlinTEEAttestationVerifier(
            _attestationVerifier
        );
        UPDATE_COOLDOWN_TIME = _updateCooldownTime;
        TRIGGER_COOLDOWN_TIME = _triggerCooldownTime;
        pcrs = _pcrs;
    }

    function updateState(
        int256 _deltaX,
        int256 _deltaY,
        uint256 _timestamp,
        uint256 _leftMuscle,
        uint256 _rightMuscle,
        uint256 _positionPrice
    ) external {
        require(enclave == msg.sender, InvalidCaller());

        wormState.leftMuscle = _leftMuscle;
        wormState.rightMuscle = _rightMuscle;
        lastUpdatedTimestamp = block.timestamp;

        emit WormStateUpdated(
            _deltaX,
            _deltaY,
            _leftMuscle,
            _rightMuscle,
            _timestamp,
            _positionPrice
        );
    }

    function updateStateByUserTrigger(
        int256 _deltaX,
        int256 _deltaY,
        uint256 _timestamp,
        uint256 _leftMuscle,
        uint256 _rightMuscle,
        address _triggeringUser
    ) external {
        require(enclave == msg.sender, InvalidCaller());

        wormState.leftMuscle = _leftMuscle;
        wormState.rightMuscle = _rightMuscle;
        lastUpdatedTimestamp = block.timestamp;

        emit WormStateUpdatedByUser(
            _deltaX,
            _deltaY,
            _leftMuscle,
            _rightMuscle,
            _timestamp,
            _triggeringUser
        );
    }

    function updateEnclaveKey(
        bytes calldata _enclaveKey,
        bytes calldata _seal,
        uint64 _timestampInMilliseconds
    ) external {
        require(
            block.timestamp > lastUpdatedTimestamp + UPDATE_COOLDOWN_TIME,
            UpdateCooldownNotOver()
        );

        ATTESTATION_VERIFIER.verify(
            _enclaveKey,
            _seal,
            _timestampInMilliseconds,
            pcrs
        );
        enclave = _pubKeyToAddress(_enclaveKey);
        lastUpdatedTimestamp = block.timestamp;

        emit EnclaveKeyUpdated(enclave);
    }

    function _pubKeyToAddress(
        bytes memory _pubKey
    ) internal pure returns (address) {
        require(_pubKey.length == 64, PubkeyLengthInvalid());

        bytes32 hash = keccak256(_pubKey);
        return address(uint160(uint256(hash)));
    }

    function trigger() external {
        require(
            block.timestamp > lastTriggeredTimestamp + TRIGGER_COOLDOWN_TIME,
            TriggerCooldownNotOver()
        );

        lastTriggeredTimestamp = block.timestamp;

        emit UserTriggeredWorm(msg.sender);
    }
}
