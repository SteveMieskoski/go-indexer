

- Need to investigate why so many receipts or transactions are not getting processed as noted in the processed value in postgres
- Also, similarly, need to identify when and flag to resolve when there is a discrepancy 
  - Probably could note the Tx hashes from the transactions, and check that each has a receipt present
  - Where either a tx without a receipt or a receipt without a tx arises, not it in the track for retries table
- Also need to verify that the retry didn't error, and if it did then a future attempt needs to be scheduled