// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * LimsHashRegistry — immutable hash registry for laboratory event integrity.
 *
 * Architecture principle: LIMS stores full laboratory data.
 * This contract stores only proof of integrity (hashes) — never patient data or results.
 *
 * Integration flow:
 *   LIMS event → Go Adapter (SHA-256) → registerHash() → blockchain
 *   LIMS current data → Go Adapter (SHA-256) → verifyHash() → integrity result
 */
contract LimsHashRegistry {

    struct HashRecord {
        string limsRecordId;
        string eventType;
        bytes32 dataHash;
        uint256 timestamp;
        string status;
        address registeredBy;
        bool exists;
    }

    // limsRecordId → HashRecord
    mapping(string => HashRecord) private records;

    // Ordered list of record IDs for enumeration
    string[] private recordIds;

    event HashRegistered(
        string limsRecordId,
        string eventType,
        bytes32 dataHash,
        uint256 timestamp,
        string status,
        address registeredBy
    );

    event HashUpdated(
        string limsRecordId,
        string eventType,
        bytes32 oldHash,
        bytes32 newHash,
        uint256 timestamp
    );

    /**
     * Register or update a hash for a LIMS laboratory event.
     * Called by the Go Adapter after calculating SHA-256 of event metadata.
     */
    function registerHash(
        string memory limsRecordId,
        string memory eventType,
        bytes32 dataHash,
        string memory status
    ) public {
        require(bytes(limsRecordId).length > 0, "LimsHashRegistry: limsRecordId required");
        require(bytes(eventType).length > 0, "LimsHashRegistry: eventType required");
        require(dataHash != bytes32(0), "LimsHashRegistry: dataHash required");

        if (!records[limsRecordId].exists) {
            recordIds.push(limsRecordId);

            records[limsRecordId] = HashRecord({
                limsRecordId: limsRecordId,
                eventType:    eventType,
                dataHash:     dataHash,
                timestamp:    block.timestamp,
                status:       status,
                registeredBy: msg.sender,
                exists:       true
            });
        } else {
            bytes32 oldHash = records[limsRecordId].dataHash;

            records[limsRecordId].eventType = eventType;
            records[limsRecordId].dataHash  = dataHash;
            records[limsRecordId].timestamp = block.timestamp;
            records[limsRecordId].status    = status;

            emit HashUpdated(
                limsRecordId,
                eventType,
                oldHash,
                dataHash,
                block.timestamp
            );
        }

        emit HashRegistered(
            limsRecordId,
            eventType,
            dataHash,
            block.timestamp,
            status,
            msg.sender
        );
    }

    /**
     * Retrieve stored hash record for a given LIMS record ID.
     * Used by the adapter to compare against a freshly computed hash.
     */
    function getHash(string memory limsRecordId)
        public
        view
        returns (
            string memory limsId,
            string memory eventType,
            bytes32 dataHash,
            uint256 timestamp,
            string memory status,
            address registeredBy
        )
    {
        require(records[limsRecordId].exists, "LimsHashRegistry: record not found");
        HashRecord storage r = records[limsRecordId];
        return (
            r.limsRecordId,
            r.eventType,
            r.dataHash,
            r.timestamp,
            r.status,
            r.registeredBy
        );
    }

    /**
     * Verify data integrity.
     * Returns true if the submitted hash matches the stored hash — no tampering detected.
     * Returns false if hashes differ — data in LIMS may have been modified.
     */
    function verifyHash(string memory limsRecordId, bytes32 dataHash)
        public
        view
        returns (bool verified)
    {
        if (!records[limsRecordId].exists) {
            return false;
        }
        return records[limsRecordId].dataHash == dataHash;
    }

    /**
     * Total number of registered hash records.
     */
    function getRecordCount() public view returns (uint256) {
        return recordIds.length;
    }

    /**
     * Get LIMS record ID by index — for enumeration by the adapter.
     */
    function getRecordIdAtIndex(uint256 index) public view returns (string memory) {
        require(index < recordIds.length, "LimsHashRegistry: index out of bounds");
        return recordIds[index];
    }
}
