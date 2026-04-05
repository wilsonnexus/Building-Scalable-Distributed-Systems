# Report Notes

## How the Leader-Follower system works

There are 5 nodes total: 1 leader and 4 followers.
All writes go to the leader.
The leader updates itself first, then replicates to followers.

### W=5, R=1
The leader waits for acknowledgements from all 4 followers before returning 201.
A read only needs the leader copy.

### W=1, R=5
The leader updates itself and returns as soon as the leader write completes.
Replication to followers continues in the background.
A read contacts all 5 nodes and returns the newest version.

### W=3, R=3
The leader returns after the leader plus 2 followers acknowledge.
A read collects 3 copies and returns the newest version.

Logical version numbers are stored with each key-value pair so the newest version can be selected.

The local_read endpoint is used only for testing. It reads the local state on a node without the leader proxy path. This helps expose the inconsistency window during propagation.

## How the Leaderless system works

There are 5 peer nodes.
Any node can receive a write request.
The node that receives the write becomes the write coordinator.
It writes locally, sends the update to the other 4 nodes, waits for all of them, and then returns 201.

A normal read returns the node's own local value.
This means a client can hit a stale node during the write propagation window.
That behavior is intentional for the assignment because it shows the inconsistency window.

## Tricky parts

- In leader-follower, W=1 returns early while replication continues, which creates a stale-read window.
- In leader-follower, followers proxy normal reads to the leader so that public reads can go to any instance.
- In leaderless, any node can act as write coordinator, so stale reads are easier to expose during propagation.
- Version numbers are necessary to compare records and return the newest value for quorum reads.

## Error handling

- Empty keys return 400.
- Missing keys return 404.
- If the required number of acknowledgements is not reached, the write returns 500.
- Internal peer failures are caught and handled so the system can still expose quorum behavior.