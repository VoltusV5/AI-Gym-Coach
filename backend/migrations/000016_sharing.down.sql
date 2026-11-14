BEGIN;
DELETE FROM sportapp.operation_results WHERE operation LIKE 'share_fork:%' OR operation LIKE 'program_fork:%';
DROP TABLE sportapp.saved_shares,sportapp.public_shares;
COMMIT;
