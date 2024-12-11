// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

interface IMarlinTEEAttestationVerifier {
    error TooOld();
    error PubkeyTooBig();

    function verify(
        bytes calldata _signerPubkey,
        bytes calldata _seal,
        uint64 _timestampInMilliseconds,
        bytes calldata _pcrs
    ) external view;
}
