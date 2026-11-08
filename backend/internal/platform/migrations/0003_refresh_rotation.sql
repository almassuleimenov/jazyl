-- 0003: distinguish tokens rotated by /auth/refresh from tokens revoked by logout, so that two
-- parallel refreshes with the same token (dashboard prefetches) are not mistaken for theft.
ALTER TABLE refresh_tokens
    ADD COLUMN rotated_at timestamptz;
