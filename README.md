## Funnel

Funnel is a go library for parsing arbitrary filter statements and building safe SQL.

This library is meant to be used as a way to build an API that accepts arbitrary filters on data.


Lets jump right into examples

Imagine this GET request:
```http request
GET /friends?query='Age >= "21" && (Drinks = "Beer" || Likes = "Sports")'
```

To convert that complex query into a SQL statement that has no chance of SQL injection is a difficult task

This library attempts to solve that use-case. 


To convert that string into a safe SQL query, with argument literals stripped out, you would simply call:
```go
allowedKeys := []string{"Age", "Drinks", "Likes"}
sqlString, sqlArgs, error := funnel.ToSql(`Age >= "21" && (Drinks = "Beer" || Likes = "Sports")`, allowedKeys) 
```

`sqlString` is represents standard sql representing a `WHERE` clause

`sqlArgs` is a `[]interface{}`

`error` is returned if the syntax of the string is incorrect, or if the string used any illegal fields

`sqlString` and `sqlArgs` are meant to be used in a sql statement, like `DB.Query(sqlString, sqlArgs...)`

### Prepared Statement
funnel uses question mark (`?`) prepared statement syntax for representing placeholders for real literals.

However, several popular databases use other syntax. To use another syntax, say Dollar `$1` format, you can just use
```go
sql = Replace(sql, funnel.Dollar)
``` 

The replace code is copy pasted from [squirrel's implementation](https://github.com/Masterminds/squirrel/blob/master/placeholder.go) so people don't have to install that dependency just for placeholders.

### Squirrel
This package is even more useful when used in conjunction with squirrel. See this snippet for an example
```go
filterSql, filterArgs, _ := funnel.ToSql(input, allowedKeys)

squirrel.Select("*").
    From("friends").
    Where("user_id", userId).
    Where(filterSql, filterArgs...).
    PlaceHolderFormat(squirrel.Dollar).
    RunWith(DB)
```