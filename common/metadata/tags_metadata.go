package metadata

import (
	"bytes"
	"math/big"
)

const WELL_KNOWN_TAG int = 0x80

const HAS_MORE_TAGS int = 0x80

const MAX_TAG_LENGTH = 0x7F

var RouteIDKey = TagsMetadataKey{WellKnownKey: RouteID, Key: ""}

type TagsMetadataKey struct {
	WellKnownKey WellKnownKey
	Key          string
}

// metadata tags
type TagsMetadata struct {
	tags map[TagsMetadataKey]string
}

func NewTagsMetadata() *TagsMetadata {
	tags := make(map[TagsMetadataKey]string)
	return &TagsMetadata{tags}
}

func (metadata *TagsMetadata) AddKeyTag(key TagsMetadataKey, val *string) {
	metadata.tags[key] = *val
}

func (metadata *TagsMetadata) AddStringTag(key string, val *string) {
	metadata.tags[TagsMetadataKey{WellKnownKey(0), key}] = *val
}

func (metadata *TagsMetadata) AddWellKnownKeyTag(key WellKnownKey, val *string) {
	metadata.tags[TagsMetadataKey{key, ""}] = *val
}

func (metadata *TagsMetadata) Encode() []byte {
	return Encode(metadata.tags)
}

func (metadata *TagsMetadata) GetRouteID() string {
	return metadata.tags[RouteIDKey]
}

func (metadata *TagsMetadata) GetByWellKnownKey(key WellKnownKey) string {
	return metadata.tags[TagsMetadataKey{key, ""}]
}

func (metadata *TagsMetadata) GetTags() map[TagsMetadataKey]string {
	return metadata.tags
}

func Encode(tags map[TagsMetadataKey]string) []byte {
	encodedByte := make([]byte, 1024)
	encodedData := bytes.NewBuffer(encodedByte)

	dataLen := len(tags)
	current := 0
	for k, val := range tags {
		current++
		if 0 != k.WellKnownKey {
			keyLength := WELL_KNOWN_TAG | int(k.WellKnownKey)
			encodedData.WriteByte(byte(keyLength))

		} else {
			keyString := k.Key
			keyStringByteArray := []byte(keyString)
			keyStringByteArrayLength := len(keyStringByteArray)
			if 0 == keyStringByteArrayLength || keyStringByteArrayLength > MAX_TAG_LENGTH {
				continue
			}
			encodedData.WriteByte(byte(keyStringByteArrayLength))
			encodedData.WriteString(keyString)
		}
		hasMoreTags := current < dataLen
		valueLength := len([]byte(val))
		// TODO probably bug just write key no value
		if 0 == valueLength || valueLength > MAX_TAG_LENGTH {
			continue
		}

		var valueByte int
		if hasMoreTags {
			valueByte = HAS_MORE_TAGS | valueLength
		} else {
			valueByte = valueLength
		}
		encodedData.WriteByte(byte(valueByte))
		encodedData.WriteString(val)
	}
	return encodedData.Bytes()
}

func EncodeBigInt(data []byte, bigInt *big.Int) []byte {
	actualByte := bigInt.Bytes()
	needAppendByteLen := 16 - len(actualByte)
	return append(append(data, make([]byte, needAppendByteLen)...), actualByte...)
}

func Decode(bytes []byte) TagsMetadata {
	return DecodeOffset(0, bytes)
}

func DecodeOffset(offset int, bytes []byte) TagsMetadata {
	tags := make(map[TagsMetadataKey]string)

	if offset >= len(bytes) {
		return TagsMetadata{tags}
	}

	for {
		keyByte := int(bytes[offset])
		offset++

		isWellKnownTag := (keyByte & WELL_KNOWN_TAG) == WELL_KNOWN_TAG
		keyLength := keyByte & MAX_TAG_LENGTH

		var key TagsMetadataKey
		if isWellKnownTag {
			key = TagsMetadataKey{WellKnownKey(keyLength), ""}
		} else {
			keyString := string(bytes[offset : offset+keyLength])
			key = TagsMetadataKey{0, keyString}
			offset = offset + keyLength
		}
		valueByte := bytes[offset]
		offset++
		hasMoreTags := (int(valueByte) & HAS_MORE_TAGS) == HAS_MORE_TAGS
		valueLength := int(valueByte & MAX_TAG_LENGTH)
		value := string(bytes[offset : offset+valueLength])

		offset = offset + valueLength

		tags[key] = value
		if !hasMoreTags {
			break
		}
	}
	return TagsMetadata{tags}
}

func DecodeBigInt(bytes []byte) *big.Int {
	id := new(big.Int).SetBytes(bytes[0:16])
	return id
}

func DecodeString(offset int, bytes []byte) (string, int) {
	length := int(bytes[offset])
	offset++
	val := string(bytes[offset : offset+length])
	return val, offset + length

}

type RouteSetup struct {
	TagsMetadata
	ID          *big.Int
	ServiceName *string
}

func (routeSetup RouteSetup) GetEnrichedTagsMetadata() TagsMetadata {
	existing := routeSetup.TagsMetadata.GetTags()
	tags := make(map[TagsMetadataKey]string)
	for k, v := range existing {
		tags[k] = v
	}
	tags[TagsMetadataKey{ServiceName, ""}] = *routeSetup.ServiceName
	tags[TagsMetadataKey{RouteID, ""}] = routeSetup.ID.String()
	tagsMetadata := TagsMetadata{tags}
	return tagsMetadata
}

func DecodeRouteSetup(bytes []byte) RouteSetup {
	id := new(big.Int).SetBytes(bytes[0:16])
	serviceName, offset := DecodeString(16, bytes)
	idStr := id.String()
	tags := DecodeOffset(offset, bytes)
	tags.AddWellKnownKeyTag(RouteID, &idStr)
	tags.AddWellKnownKeyTag(ServiceName, &serviceName)
	setup := RouteSetup{ID: id, ServiceName: &serviceName, TagsMetadata: tags}
	// tagsMetadata.addWellKnownKeyTag(RouteID, &id)
	// tagsMetadata.addWellKnownKeyTag(ServiceName, &serviceName)
	return setup
}

type Forwarding struct {
	TagsMetadata
	OriginRouteID *big.Int
}

func DecodeForwarding(bytes []byte) Forwarding {
	id := new(big.Int).SetBytes(bytes[0:16])

	tags := DecodeOffset(16, bytes)
	forwarding := Forwarding{OriginRouteID: id, TagsMetadata: tags}
	// tagsMetadata.addWellKnownKeyTag(RouteID, &id)
	// tagsMetadata.addWellKnownKeyTag(ServiceName, &serviceName)
	return forwarding
}
