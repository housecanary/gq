package ts

import "github.com/housecanary/gq/types"

var BuiltinTypes = NewModule()

var IDType = NewScalarType[types.ID](BuiltinTypes, "")
var StringType = NewScalarType[types.String](BuiltinTypes, "")
var IntType = NewScalarType[types.Int](BuiltinTypes, "")
var FloatType = NewScalarType[types.Float](BuiltinTypes, "")
var BooleanType = NewScalarType[types.Boolean](BuiltinTypes, "")
