// SPDX-License-Identifier: MIT

pragma solidity ^0.8.28;

import {IMarlinTEEAttestationVerifier} from "./IMarlinTEEAttestationVerifier.sol";

contract Worm {
    IMarlinTEEAttestationVerifier public immutable ATTESTATION_VERIFIER;
    uint256 public immutable COOLDOWN_TIME;

    uint256 public lastUpdatedTimestamp;
    uint256 public lastTriggeredTimestamp;
    address public enclave;
    bytes public pcrs;
    WormState private wormState;

    event WormStateUpdated(
        uint256 deltaX,
        uint256 deltaY,
        uint256 positionPrice,
        uint256 leftMuscle,
        uint256 rightMuscle,
        uint256 positionTimestamp
    );
    event WormStateUpdatedByUser(
        address indexed triggeringUser,
        uint256 deltaX,
        uint256 deltaY,
        uint256 leftMuscle,
        uint256 rightMuscle,
        uint256 positionTimestamp
    );
    event EnclaveKeyUpdated(bytes indexed enclaveKey);
    event UserTriggeredWorm(address indexed triggeringUser);

    error PubkeyLengthInvalid();
    error UpdateCooldownNotOver();
    error TriggerCooldownNotOver();
    error InvalidCaller();

    struct WormState {
        uint256 leftMuscle;
        uint256 rightMuscle;
    }

    constructor(address _attestationVerifier, uint256 _cooldownTime, bytes memory _pcrs) payable {
        ATTESTATION_VERIFIER = IMarlinTEEAttestationVerifier(_attestationVerifier);
        COOLDOWN_TIME = _cooldownTime;
        pcrs = _pcrs;
    }

    function updateState(
        uint256 _deltaX,
        uint256 _deltaY,
        uint256 _positionTimestamp,
        uint256 _positionPrice,
        uint256 _leftMuscle,
        uint256 _rightMuscle
    ) public {
        require(enclave == msg.sender, InvalidCaller());

        wormState.leftMuscle = _leftMuscle;
        wormState.rightMuscle = _rightMuscle;

        emit WormStateUpdated(_deltaX, _deltaY, _positionPrice, _leftMuscle, _rightMuscle, _positionTimestamp);
    }

    function updateStateByUserTrigger(
        uint256 _deltaX,
        uint256 _deltaY,
        uint256 _positionTimestamp,
        uint256 _leftMuscle,
        uint256 _rightMuscle,
        address _triggeringUser
    ) public {
        require(enclave == msg.sender, InvalidCaller());

        wormState.leftMuscle = _leftMuscle;
        wormState.rightMuscle = _rightMuscle;

        emit WormStateUpdatedByUser(_triggeringUser, _deltaX, _deltaY, _leftMuscle, _rightMuscle, _positionTimestamp);
    }

    function updateEnclaveKey(bytes calldata _enclaveKey, bytes calldata _seal, uint64 _timestampInMilliseconds)
        public
    {
        require(block.timestamp > lastUpdatedTimestamp + COOLDOWN_TIME, UpdateCooldownNotOver());

        ATTESTATION_VERIFIER.verify(_enclaveKey, _seal, _timestampInMilliseconds, pcrs);
        enclave = _pubKeyToAddress(_enclaveKey);
        lastUpdatedTimestamp = block.timestamp;

        emit EnclaveKeyUpdated(_enclaveKey);
    }

    function _pubKeyToAddress(bytes memory _pubKey) internal pure returns (address) {
        require(_pubKey.length == 64, PubkeyLengthInvalid());

        bytes32 hash = keccak256(_pubKey);
        return address(uint160(uint256(hash)));
    }

    function trigger() external {
        // can only trigger every 60 seconds
        require(lastTriggeredTimestamp + 60 < block.timestamp, TriggerCooldownNotOver());

        lastTriggeredTimestamp = block.timestamp;

        emit UserTriggeredWorm(msg.sender);
    }

    function getWormState() external view returns (uint256, uint256) {
        return (wormState.leftMuscle, wormState.rightMuscle);
    }
}
