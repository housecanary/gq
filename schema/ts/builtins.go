package ts

import "github.com/housecanary/gq/types"

var BuiltinTypes = Module()

var IDType = Scalar[types.ID](BuiltinTypes, "")
var StringType = Scalar[types.String](BuiltinTypes, "")
var IntType = Scalar[types.Int](BuiltinTypes, "")
var FloatType = Scalar[types.Float](BuiltinTypes, "")
var BooleanType = Scalar[types.Boolean](BuiltinTypes, "")
