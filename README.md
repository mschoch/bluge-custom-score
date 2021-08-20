## bluge-custom-score


This example application attempts to illustrate the following:

 - Documents are indexed with a field named 'boost', this field is only stored, not indexed.
 - The value of the boost field is a 64-bit floating point number, and we use Go standard-library functions to read/write these bytes
 - Our goal is to run a particular search twice
 - First, we run the search normally
 - Second, we wrap the search query with a custom query implementation
 - This custom query implementation matches the same set of documents as the original query
 - But, each document match's score is transformed with a custom function
 - This user-provided function can load document stored fields using the same reader as the search.

In summary, this example illustrates how you can perform an arbitrary scoring transformation (at any point in the query hierarchy) and use both the original score and document contents in the function.
