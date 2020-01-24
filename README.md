# GQ - A Graph Query library for go

GQ is a library for implementing GraphQL servers

## Why?

There are several projects that implement GraphQL servers for Go:
  * [gqlgen](https://github.com/99designs/gqlgen)
  * [gophers](https://github.com/graph-gophers/graphql-go)
  * [graphql-go](https://github.com/graphql-go/graphql)
  * [thunder](https://github.com/samsarahq/thunder)

This project aims to improve on existing libraries in two key areas: Ease of schema creation and batched data loading.

### Schema creation

Existing approaches make schema creation and maintenance more difficult than it could be.

#### gqlgen

Schema first approach. You write a schema, and then bind types to the schema
in a config file, then use the schema and config file to generate code. Weakness - need to keep definitions in schema and structs in sync, requires code generation.

#### gophers

Schema first approach. You write a schema, and then provide a root type to bind to the schema. Methods are bound to the schema. Weakness - you have to write a method for each field, leading to lots of boilerplate for large DTOs.

#### graphql-go

Schema in code. You build up the schema by constructing it in Go objects. Weakness - very verbose/boilerplate heavy, not type checked.

#### thunder

Schema built from structs, with separate method registration. Weakness - verbose registration of resolver methods. Limited flexibility in schema definition.

#### GQ

GQ takes a hybrid approach to schema definition. Although you can build a schema "by hand" in a way very similar to graphql-go, the preferred approach is to define the schema via annotated structs. GQ attempts to infer reasonable defaults via reflection on the structs, but the full power of GQL schema definitions is available in struct tags to customize the schema (i.e. add directives, specify exact return types, etc). The GQL fragments are attached to the types they are associated with, so all data about the type is in one place, not split across many locations. See below for examples of this approach.

### Data loading

A common use case for GraphQL servers is loading data from other systems (databases or internal services) and exposing this in a schema that clients can easily use. Major performance gains can be had by batching together related queries to reduce the overhead of each query.

Existing libraries all make it difficult to optimally batch queries: there is no hook to indicate good batch points, so the libraries rely on timeouts (i.e. every 10 msec batch up all pending queries and run them). This leads to unnecessary latency (i.e. if batch timeout = 20ms and the query has a depth of 5 up to 100ms of time is wasted) or suboptimal batching (i.e. setting a lower timeout, which then may not include all queries).

Event loop based systems like the Node and Python implementations solve this relatively easily:  they simply schedule batches on the next event loop tick. The net result of this is that batches are issued ASAP once all synchronous resolvers are executed.

GQ attempts to mimic the behavior of event loop based systems by notifying a listener when all resolvers are blocked. The listener can use this notification to batch and send pending queries to backend systems.

In addition, GQ's model allows resolvers and data loaders to control the number of goroutines created, while other systems simply wrap each non-trivial resolver in a separate goroutine, leading to an explosion of goroutines for larger queries.

## Supported Features

| | |
| --------: | :-------- |
| Kind | struct first |
| Boilerplate | less |
| Query | :+1: |
| Mutation | :no_entry: |
| Subscription | :no_entry: |
| Type Safety | :+1: |
| Type Binding | :+1: |
| Embedding | :+1: |
| Interfaces | :+1: |
| Generated Enums | :+1: |
| Generated Inputs | :+1: |
| Stitching gql | :no_entry: |
| Opentracing | :no_entry: |
| Hooks for error logging | :+1: |
| Dataloading | :+1: |
| Concurrency | :+1: |

## Limitations

Currently the biggest imitations of GQ are no support for Mutations or Subscriptions. Mutations would be relatively easy to add in the existing model, and subscriptions should be possible. Neither has been implemented since the author's use case does not require them.

Another limitation of GQ is that query parsing is relatively slow. GQ uses an ANTLR based parser that allocates and type checks a good deal while parsing queries. This is an implementation detail that could be changed; however, GQ does include support for caching and reusing query execution plans. If your use case includes a fixed (or slowly changing) set of queries executed by clients, this completely negates the slowness of parsing a query, since the query is parsed and prepared once, and then executed many times.

## Usage

See the godoc examples, and the examples in the `examples` folder for comprehensive examples. A brief overview is presented here.

### Schema definition

Schemas in GQ are defined by annotating structs (manual schema definition is also possible, see introspection.go in the schema package, or the query tests for examples). See the `schema/structschema` docs for a complete description of the options.

#### Object Types

Object types are defined as a plain struct. This struct may include a structschema.Meta field that has a struct field tag containing a GQL schema definition of that object type. The GQL will be merged with reflection data to produce the final schema. Fields declared in the GQL that do not correspond to fields of the struct are expected to have a resolver method of the form `Resolve<FieldName>([arg1 ... argN]) <Return Type>`. Fields of the struct may also be customised by adding GQL definitions of the field

Example:

```go
type Human struct {
    structschema.Meta `
        "A Human contains information about a humanoid person"
        {
            """
            Returns the friends of degree N
            (i.e. degree 1 - direct friends,
            degree 2 - friends of friends, etc)
            """
            friends(degree: Int = 1): [Person!]
        }
    `
    Name       String `gq:":String!;The name of the person"`
    BestFriend String `gq:"best @deprecated"`
    Age        Int
    Birthday   Date
    Password   String `gq:"-"`
}

func (h *Human) ResolveFriends(degree Int) []*Person {
    ...
}
```

Would produce the final schema definition
```

"A Human contains information about a humanoid person"
object Human {
    """
    Returns the friends of degree N
    (i.e. degree 1 - direct friends,
    degree 2 - friends of friends, etc)
    """
    friends(degree: Int = 1): [Person!]

    "The name of the person"
    name: String!

    best: String @deprecated

    age: Int

    birthday: Date
}

```

#### Interface types

Interface types are defined as a struct with a single `Interface` field. The `Interface` field should be of type `interface{...}` and define any methods needed on implementations of the interface. A struct field tag defining fields of the interface can be attached to the `Interface` field.

Example:
```go
type Named struct {
    Interface interface {
        isNamed()
    } `{
        name: String!
    }`
}

type Pet struct {
    Name String
}

func (Pet) isNamed()

type Human struct {
    Name String
}

func (Human) isNamed()
```

#### Union types

Interface types are defined as a struct with a single `Union` field. The `Union` field should be of type `interface{...}` and define any methods needed on members of the union. A struct field tag with a description or directives to apply to the union can be attached to the field.

Example:
```go
type PetOrHuman struct {
    Union interface {
        isPetOrHuman()
    } `"This is a union" @foo`
}

type Pet struct {
    ...
}

func (Pet) isPetOrHuman()

type Human struct {
    ...
}

func (Human) isPetOrHuman()
```

#### Enum types

Enum types are defined as a struct with an embedded structschema.Enum field. The tag on the Enum field defines the interface values and type definition.

Example:

```go
type Drinks struct {
    structschema.Enum `"Types of drinks you can order" {
        "A sweet carbonated drink."
        SOFTDRINK

        "Brewed tea leaves. Iced or not, sweet or not"
        TEA

        "Brewed coffee beans"
        COFFEE @neededForCoding

        "Fermented grain drink"
        BEER

        "Fermented grape drink"
        WINE
    }`
}
```

#### Scalar types

Anything that implements `types.ScalarMarshaler` and `types.ScalarUnmarshaller`. Built in scalars are available in the types package. Nothing forces the use of these built in scalars - you could define your own `String` scalar for example if you wanted to.

#### Input Objects

InputObject types are defined as a struct with an embedded structschema.InputObject field. The tag on the InputObject field defines any additional GQL data, and fields are processed like an Object.

If the input object defines a `Validate() error` method, it will be called when the input object is constructed.

Example:

```go
type NameInput struct {
    structschema.InputObject `"A full name"`
    First  String
    Middle String
    Last   String
}

func (n *NameInput) Validate() error {
    ...
}
```

### Querying

There are two steps to querying a schema. First, you prepare a query:
```go
queryText := `
    query a { foo }
    query b { bar }
`

q, err := query.PrepareQuery(queryText, /* operationName: */ a, schema)
````

then you can execute the query any number of times. To execute the query, you pass a `context.Context`, a root object, a `Variables` instance and a `QueryListener`.

The `context.Context` is made available to resolvers for their use.

The root object is used to resolve the root fields against. It should be an instance of the type defined as the Query type in the schema.

The variables is a dict of variables available to the query.

The query listener is a `query.ExecutionListener` that can be used to log execution of the query, and schedule loads at idle points.

Example:
```go
vars := query.NewVariablesFromJSON(jsonDataFromSomewhere)
response := q.Execute(context.Background(), &RootObject{}, vars, nil)
```

### Data loading and asynchronous resolvers

Many resolver methods will want to schedule asynchronous work. The model for this in GQ is that on invocation the resolver will schedule work, and then return a value that can be awaited to collect the results. GQ will schedule the await after all executable resolvers have run.

If using the `structschema` package, async resolvers should follow one of these patterns:

```go
func (t *SomeType) ResolveFoo(loader *someLoaderType) <-chan FooType {
    var c <-chan FooType
    c = loader.enqueueLoadFoo(t.id)
    return c
}
```

```go
func (t *SomeType) ResolveFoo(loader *someLoaderType) func() (FooType, error) {
    var c <-chan BarOrError
    c = loader.enqueueLoadFoo(t.id)
    return func () (FooType, error) {
        result := <- c
        if result.Error != nil {
            return nil, result.Error
        }

        return transformBarToFoo(result.Bar), nil
    }
}
```

The best way to schedule batch loads is to enqueue a request in the resolver, and in a `QueryListener` trigger batch loads of all pending enqueued requests when `NotifyIdle` is called.

Note that a resolver should never block the caller:  instead, it should return a value that the caller can use to await the result when convenient - either a callback function to produce the final result, or a channel.
