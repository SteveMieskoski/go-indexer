


- Too many collisions with tx hashes when adding to the DB after pulled from kafka.  
  - Expect duplicates are getting added or the code is trying to add them multiple times.
    - This is occurring primarily with the Receipt and log repositories