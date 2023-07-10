CREATE TABLE identity_registration_codes
(
    id CHAR(36) NOT NULL PRIMARY KEY,
    code VARCHAR (64) NOT NULL, -- HMACed value of the actual code
    used_at timestamp NULL DEFAULT NULL,
    code_type int NOT NULL,
    expires_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    issued_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    selfservice_registration_flow_id CHAR(36),
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    nid CHAR(36) NOT NULL,
    identity_id CHAR(36) NOT NULL,
    CONSTRAINT identity_registration_codes_selfservice_registration_flows_id_fk
        FOREIGN KEY (selfservice_registration_flow_id)
        REFERENCES selfservice_registration_flows (id)
        ON DELETE cascade,
    CONSTRAINT identity_registration_codes_identity_id_fk
        FOREIGN KEY (identity_id)
        REFERENCES identities (id)
        ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT identity_registration_codes_networks_id_fk
        FOREIGN KEY (nid)
        REFERENCES networks (id)
        ON UPDATE RESTRICT ON DELETE CASCADE
);

CREATE INDEX identity_registration_codes_nid_flow_id_idx ON identity_registration_codes (nid, selfservice_registration_flow_id);
CREATE INDEX identity_registration_codes_id_nid_idx ON identity_registration_codes (id, nid);

ALTER TABLE selfservice_registration_flows ADD submit_count int NOT NULL DEFAULT 0;
ALTER TABLE selfservice_registration_flows ADD skip_csrf_check boolean NOT NULL DEFAULT FALSE;
