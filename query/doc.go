// Copyright 2018 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package query provides all the code to prepare and execute a GraphQL query.
//
// To execute a query one needs three things:  a root object, a query operation
// and a schema.  These are used as follows:
//   - root object - provides the root object the query operation will apply to.
//   - query operation - specifies a GraphQL query to execute.  The query contains
//     a description of what data is selected.
//   - schema - provides the model of the data the query is written against
//
// Design notes:
//   - No reflection allowed in this package.  All things that would require reflection should
//     happen via operations defined on types in the schema.  For example, instead of using reflection
//     to iterate over a slice, the list selector uses the schema.ListValue interface to perform the
//     iteration.  This allows resolvers to be implemented using reflection or not.
//
//   - Threading model is one goroutine per query execution.  A resolver may return an asynchronous value (see below)
//     but the results are collected on the query processor thread.  This means that none of the structures
//     in this package need to be concerned with multi-threaded access:  queries are built once, and never
//     mutated, and a single execution of a query happens on a single thread.
//
//   - Resolution model is a depth-first walk of the query tree, stopping at asynchronous nodes.
//
//     An asynchronous node is one where the resolver returns a schema.AsyncValue object.  In this case, the
//     resolver is expected to have enqueued work to be done in a separate goroutine (or, equivalently, signaled an
//     outside system to  begin a job) prior to returning the AsyncValue.  The AsyncValue represents a handle to the
//     work that was enqueued.  Invoking the Await method on the AsyncValue will initiate a blocking wait for the
//     results of the computation.
//
//     When an async node is encountered, the system enqueues a callback to await the results at a later time, and
//     proceeds with the next non-child node in the walk.  After the tree has been fully walked in this manner,
//     NotifyIdle is called on the query listener associated with the execution, and each AsyncValue is awaited.  As
//     the results arrive, depth first walking in the same manner as above is initiated on each sub-tree rooted at a
//     newly resolved async value.
//
//     As an example, assume the following query
//        lookupFooAsync {
//          a {
//            bAsync {
//              c
//              d
//            }
//          }
//          e {
//            f {
//              gAsync {
//                h
//                i
//              }
//            }
//          }
//        }
//     The order of operations to process this query would be:
//        resolve lookupFooAsync
//        notify idle
//        await lookupFooAsync
//        resolve a
//        resolve bAsync
//        resolve e
//        resolve f
//        resolve gAsync
//        notify idle
//        await bAsync
//        await gAsync
//        resolve c
//        resolve d
//        resolve h
//        resolve i
//
//     The goal of this approach is to allow efficient batching of queries to data stores.  The NotifyIdle
//     callback is a perfect place to batch schedule requests to data stores.  Scheduling work in NotifyIdle is
//     roughly equivalent to scheduling on next tick in an event loop based system.
package query
