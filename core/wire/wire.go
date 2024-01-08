package wire

import (
	"bytes"
	"fmt"

	"github.com/darlean-io/darlean.go/utils/fastproto"
)

type TransportTags struct {
	Transport_Receiver string
	Transport_Return   string
}

type RemoteCallTags struct {
	// "call" | "return"
	Remotecall_Kind string
	Remotecall_Id   string
}

type ActorCallRequest struct {
	Lazy       bool
	ActorType  string
	ActorId    []string
	ActionName string
	Arguments  []any
}

type ActorCallResponse struct {
	Error any
	Value any
}

type Tags struct {
	TransportTags
	RemoteCallTags
	ActorCallRequest
	ActorCallResponse
}

const CHAR_CODE_VERSION_MAJOR = '0'
const CHAR_CODE_VERSION_MINOR = '0'

const CHAR_CODE_RETURN = 'r'
const CHAR_CODE_CALL = 'c'

const CHAR_CODE_FALSE = 'f'
const CHAR_CODE_TRUE = 't'

func Serialize(buf *bytes.Buffer, tags Tags) error {
	// Version
	fastproto.WriteChar(buf, CHAR_CODE_VERSION_MAJOR)
	fastproto.WriteChar(buf, CHAR_CODE_VERSION_MINOR)

	// Transport
	fastproto.WriteString(buf, &tags.Transport_Receiver)
	fastproto.WriteString(buf, &tags.Transport_Return)

	// Transport failure code + message
	fastproto.WriteString(buf, nil)
	fastproto.WriteString(buf, nil)

	// Tracing cids + parentuid
	err := fastproto.WriteVariant(buf, nil)
	if err != nil {
		return err
	}
	err = fastproto.WriteString(buf, nil)
	if err != nil {
		return err
	}

	// RemoteCall
	fastproto.WriteString(buf, &tags.Remotecall_Id)
	if tags.Remotecall_Kind == "return" {
		fastproto.WriteChar(buf, CHAR_CODE_RETURN)
	} else {
		fastproto.WriteChar(buf, CHAR_CODE_CALL)
	}

	// Call request
	if tags.Lazy {
		fastproto.WriteChar(buf, CHAR_CODE_TRUE)
	} else {
		fastproto.WriteChar(buf, CHAR_CODE_FALSE)
	}
	fastproto.WriteString(buf, &tags.ActorType)
	fastproto.WriteString(buf, &tags.ActionName)
	fastproto.WriteUnsignedInt(buf, len(tags.ActorId))
	for _, part := range tags.ActorId {
		fastproto.WriteString(buf, &part)
	}
	fastproto.WriteUnsignedInt(buf, len(tags.Arguments))
	for _, arg := range tags.Arguments {
		err := fastproto.WriteVariant(buf, arg)
		if err != nil {
			return err
		}
	}

	// Call response
	fastproto.WriteVariant(buf, tags.ActorCallResponse.Value)
	fastproto.WriteJson(buf, tags.ActorCallResponse.Error)
	return nil
}

func Deserialize(buf *bytes.Buffer, tags *Tags) error {
	// Version number major + minor
	major, err := fastproto.ReadChar(buf)
	if err != nil {
		return err
	}
	if major > CHAR_CODE_VERSION_MAJOR {
		return fmt.Errorf("wire: invalid major version: %v", major)
	}

	_, err = fastproto.ReadChar(buf)
	if err != nil {
		return err
	}

	// Transport receiver + return
	receiver, err := fastproto.ReadString(buf)
	if err != nil {
		return err
	}
	tags.Transport_Receiver = *receiver

	returnTo, err := fastproto.ReadString(buf)
	if err != nil {
		return err
	}
	tags.Transport_Return = *returnTo

	// Transport failure code + message
	_, err = fastproto.ReadString(buf)
	if err != nil {
		return err
	}

	_, err = fastproto.ReadString(buf)
	if err != nil {
		return err
	}

	// Tracing cids + parentuid
	_, err = fastproto.ReadVariant(buf)
	if err != nil {
		return err
	}

	_, err = fastproto.ReadString(buf)
	if err != nil {
		return err
	}

	// Remote call
	remoteCallId, err := fastproto.ReadString(buf)
	if err != nil {
		return err
	}
	tags.Remotecall_Id = *remoteCallId

	remoteCallKind, err := fastproto.ReadChar(buf)
	if err != nil {
		return err
	}
	if remoteCallKind == CHAR_CODE_RETURN {
		tags.Remotecall_Kind = "return"
	} else {
		tags.Remotecall_Kind = "call"
	}

	// Call request
	lazy, err := fastproto.ReadChar(buf)
	if err != nil {
		return err
	}
	if lazy == CHAR_CODE_FALSE {
		tags.Lazy = false
	} else {
		tags.Lazy = true
	}

	actorType, err := fastproto.ReadString(buf)
	if err != nil {
		return err
	}
	tags.ActorType = *actorType

	actionName, err := fastproto.ReadString(buf)
	if err != nil {
		return err
	}
	tags.ActionName = *actionName

	nrIdFields, err := fastproto.ReadUnsignedInt(buf)
	if nrIdFields > 0 {
		parts := make([]string, nrIdFields)
		for i := 0; i < int(nrIdFields); i++ {
			idPart, err := fastproto.ReadString(buf)
			if err != nil {
				return err
			}
			parts[i] = *idPart
		}
		tags.ActorId = parts
	}

	nrArguments, err := fastproto.ReadUnsignedInt(buf)
	if nrArguments > 0 {
		args := make([]any, nrArguments)
		for i := 0; i < int(nrArguments); i++ {
			arg, err := fastproto.ReadVariant(buf)
			if err != nil {
				return err
			}
			args[i] = arg
		}
		tags.Arguments = args
	}

	// Call response
	responseValue, err := fastproto.ReadVariant(buf)
	if err != nil {
		return err
	}
	tags.Value = responseValue

	responseError, err := fastproto.ReadJson(buf)
	if err != nil {
		return err
	}
	tags.Error = responseError

	return nil
}
