package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	dIn := reflect.ValueOf(data)
	dOut := reflect.ValueOf(out)
	if dOut.Kind() != reflect.Ptr {
		return fmt.Errorf("out not a pointer")
	}
	dOut = reflect.ValueOf(out).Elem()

	if dOut.Kind() == reflect.Int {
		if !dIn.CanFloat() {
			return fmt.Errorf("data not an int")
		}
		dOut.SetInt(int64(dIn.Float()))
		return nil
	}

	if dOut.Kind() == reflect.String {
		if !dIn.CanConvert(dOut.Type()) {
			return fmt.Errorf("data not a string")
		}
		dOut.SetString(dIn.String())
		return nil
	}

	if dOut.Kind() == reflect.Bool {
		if !dIn.CanConvert(dOut.Type()) {
			return fmt.Errorf("data not a bool")
		}
		dOut.SetBool(dIn.Bool())
		return nil
	}

	if dOut.Kind() == reflect.Slice {
		if dIn.Kind() != reflect.Slice {
			return fmt.Errorf("data not an array")
		}
		tmpSlice := reflect.MakeSlice(dOut.Type(), dIn.Len(), dIn.Len())
		for i := 0; i < dIn.Len(); i++ {
			elemPtr := tmpSlice.Index(i).Addr().Interface()
			inputPtr := dIn.Index(i).Interface()
			err := i2s(inputPtr, elemPtr)
			if err != nil {
				return err
			}
		}
		dOut.Set(tmpSlice)
		return nil
	}

	if dOut.Kind() == reflect.Struct {
		for i := 0; i < dOut.NumField(); i++ {
			field := dOut.Field(i).Addr().Interface()
			if dIn.Kind() != reflect.Map {
				return fmt.Errorf("data not a map")
			}
			dMap := dIn.MapIndex(reflect.ValueOf(dOut.Type().Field(i).Name))
			if dMap.IsValid() {
				err := i2s(dMap.Interface(), field)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	return nil
}
