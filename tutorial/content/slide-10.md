## Select DISTINCT 

The DISTINCT keyword can be used to eliminate duplicates from the
output.

The query on the right uses the DISTINCT keyword in the SELECT
statement to produce a set of unique results.

Try removing the DISTINCT keyword from the query to see the
difference.

<pre id="example">
    SELECT DISTINCT orderlines[0].productId
        FROM orders
</pre>
