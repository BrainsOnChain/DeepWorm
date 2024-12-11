// SPDX-License-Identifier: MIT

pragma solidity ^0.8.28;

import {IMarlinTEEAttestationVerifier} from "./IMarlinTEEAttestationVerifier.sol";

contract Worm {
    IMarlinTEEAttestationVerifier public immutable ATTESTATION_VERIFIER;
    uint256 public immutable COOLDOWN_TIME;

    Position private position;
    uint256 public lastUpdatedTimestamp;
    address public enclave;
    bytes public pcrs;

    event WormPositionUpdated(uint256 x, uint256 y);
    event EnclaveKeyUpdated(bytes indexed enclaveKey);
    event UserTriggeredChemotaxis();
    event UserTriggeredNoseTouch();

    error PubkeyLengthInvalid();
    error CooldownNotOver();
    error InvalidCaller();

    struct Position {
        uint256 x;
        uint256 y;
    }

    constructor(
        address _attestationVerifier,
        uint256 _cooldownTime,
        bytes memory _pcrs
    ) payable {
        ATTESTATION_VERIFIER = IMarlinTEEAttestationVerifier(
            _attestationVerifier
        );
        COOLDOWN_TIME = _cooldownTime;
        pcrs = _pcrs;
    }

    function updateCoordinate(uint256 _x, uint256 _y) public {
        require(enclave == msg.sender, InvalidCaller());

        position.x = _x;
        position.y = _y;

        emit WormPositionUpdated(_x, _y);
    }

    function updateEnclaveKey(
        bytes calldata _enclaveKey,
        bytes calldata _seal,
        uint64 _timestampInMilliseconds
    ) public {
        require(
            block.timestamp > lastUpdatedTimestamp + COOLDOWN_TIME,
            CooldownNotOver()
        );

        ATTESTATION_VERIFIER.verify(
            _enclaveKey,
            _seal,
            _timestampInMilliseconds,
            pcrs
        );
        enclave = _pubKeyToAddress(_enclaveKey);
        lastUpdatedTimestamp = block.timestamp;

        emit EnclaveKeyUpdated(_enclaveKey);
    }

    function _pubKeyToAddress(
        bytes memory _pubKey
    ) internal pure returns (address) {
        if (!(_pubKey.length == 64)) revert PubkeyLengthInvalid();

        bytes32 hash = keccak256(_pubKey);
        return address(uint160(uint256(hash)));
    }

    function triggerChemotaxis() external {
        emit UserTriggeredChemotaxis();
    }

    function triggerNoseTouch() external {
        emit UserTriggeredNoseTouch();
    }

    function getWormPosition() external view returns (uint256 x, uint256 y) {
        x = position.x;
        y = position.y;
    }
}
