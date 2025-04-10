package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type ObjectGet struct {
	object  object.Object
	payload []byte
}

func findObjectByFilePath(objects []ObjectGet, targetPath string) (ObjectGet, bool) {
	for _, objGet := range objects {
		attrs := objGet.object.Attributes()
		for _, attr := range attrs {
			if attr.Key() == object.AttributeFilePath && attr.Value() == targetPath {
				return objGet, true
			}
		}
	}
	return ObjectGet{}, false
}

func uploadObjectFromFilePath(root string, filePath string, pool pool.Pool, ctx context.Context, objects []ObjectGet) {
	relPath, err := filepath.Rel(root, filePath)
	panicOnErr(err)
	fmt.Println("Uploading file:", relPath)

	objFound, found := findObjectByFilePath(objects, relPath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if found {
		if reflect.DeepEqual(content, objFound.payload) {
			fmt.Println("File already uploaded, skipping...")
			return
		} else {
			fmt.Println("File content changed, replacing...")
			var prmObjectDelete client.PrmObjectDelete
			_, err := pool.ObjectDelete(ctx, containerId(), objFound.object.GetID(), signer(), prmObjectDelete)
			panicOnErr(err)
		}
	}

	attrs := []object.Attribute{
		*object.NewAttribute(object.AttributeTimestamp, strconv.FormatInt(time.Now().Unix(), 10)),
		*object.NewAttribute(object.AttributeFileName, filepath.Base(relPath)),
		*object.NewAttribute(object.AttributeFilePath, relPath),
		*object.NewAttribute(object.AttributeContentType, mime.TypeByExtension(filepath.Ext(relPath))),
	}
	var obj object.Object
	obj.SetContainerID(containerId())
	obj.SetOwner(signer().UserID())
	obj.SetAttributes(attrs...)

	var prmObjectPutInit client.PrmObjectPutInit
	objWriter, err := pool.ObjectPutInit(ctx, obj, signer(), prmObjectPutInit)
	panicOnErr(err)

	_, err = objWriter.Write(content)
	if err != nil {
		_ = objWriter.Close()
		panicOnErr(err)
	}

	err = objWriter.Close()
	panicOnErr(err)
}

func getAllObjectsFromContainer(signer user.Signer, pool pool.Pool, ctx context.Context) []ObjectGet {
	var prmObjectSearch client.PrmObjectSearch
	objectIdsArray, err := pool.ObjectSearchInit(ctx, containerId(), signer, prmObjectSearch)
	panicOnErr(err)

	var objects []ObjectGet

	var prmObjectGet client.PrmObjectGet

	objectIdsArray.Iterate(func(i oid.ID) bool {
		obj, payloadReader, err := pool.ObjectGetInit(ctx, containerId(), i, signer, prmObjectGet)
		panicOnErr(err)

		payload, err := io.ReadAll(payloadReader)
		panicOnErr(err)

		objGet := ObjectGet{
			object:  obj,
			payload: payload,
		}
		objects = append(objects, objGet)
		return false
	})

	return objects
}

func uploadDir(folderPath string) {
	files, err := findFilesRecursively(folderPath)
	panicOnErr(err)

	signer := signer()
	nodeParams := pool.NewFlatNodeParams(storageNodes())

	poolOptions := pool.DefaultOptions()
	poolOptions.SetNodeStreamTimeout(time.Second * 15)
	pool, err := pool.New(nodeParams, signer, poolOptions)
	if err != nil {
		pool.Close()
		panicOnErr(err)
	}

	ctx := context.Background()

	err = pool.Dial(ctx)
	defer pool.Close()
	panicOnErr(err)

	fmt.Println("Getting all objects from container...")
	objects := getAllObjectsFromContainer(signer, *pool, ctx)

	for _, file := range files {
		uploadObjectFromFilePath(folderPath, file, *pool, ctx, objects)
	}
}
