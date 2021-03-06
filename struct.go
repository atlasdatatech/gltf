package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Index is an utility function that returns a pointer to a uint32.
func Index(i uint32) *uint32 {
	return &i
}

// Float64 is an utility function that returns a pointer to a float64.
func Float64(val float64) *float64 {
	return &val
}

// An Asset is metadata about the glTF asset.
type Asset struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Copyright  string      `json:"copyright,omitempty"`         // A copyright message suitable for display to credit the content creator.
	Generator  string      `json:"generator,omitempty"`         // Tool that generated this glTF model. Useful for debugging.
	Version    string      `json:"version" validate:"required"` // The glTF version that this asset targets.
	MinVersion string      `json:"minVersion,omitempty"`        // The minimum glTF version that this asset targets.
}

// Document defines the root object for a glTF asset.
type Document struct {
	Extensions         Extensions   `json:"extensions,omitempty"`
	Extras             interface{}  `json:"extras,omitempty"`
	ExtensionsUsed     []string     `json:"extensionsUsed,omitempty"`
	ExtensionsRequired []string     `json:"extensionsRequired,omitempty"`
	Accessors          []Accessor   `json:"accessors,omitempty" validate:"dive"`
	Animations         []Animation  `json:"animations,omitempty" validate:"dive"`
	Asset              Asset        `json:"asset"`
	Buffers            []Buffer     `json:"buffers,omitempty" validate:"dive"`
	BufferViews        []BufferView `json:"bufferViews,omitempty" validate:"dive"`
	Cameras            []Camera     `json:"cameras,omitempty" validate:"dive"`
	Images             []Image      `json:"images,omitempty" validate:"dive"`
	Materials          []Material   `json:"materials,omitempty" validate:"dive"`
	Meshes             []Mesh       `json:"meshes,omitempty" validate:"dive"`
	Nodes              []Node       `json:"nodes,omitempty" validate:"dive"`
	Samplers           []Sampler    `json:"samplers,omitempty" validate:"dive"`
	Scene              *uint32      `json:"scene,omitempty"`
	Scenes             []Scene      `json:"scenes,omitempty" validate:"dive"`
	Skins              []Skin       `json:"skins,omitempty" validate:"dive"`
	Textures           []Texture    `json:"textures,omitempty" validate:"dive"`
}

// An Accessor is a typed view into a bufferView.
// An accessor provides a typed view into a bufferView or a subset of a bufferView
// similar to how WebGL's vertexAttribPointer() defines an attribute in a buffer.
type Accessor struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	Name          string        `json:"name,omitempty"`
	BufferView    *uint32       `json:"bufferView,omitempty"`
	ByteOffset    uint32        `json:"byteOffset,omitempty"`
	ComponentType ComponentType `json:"componentType" validate:"lte=5"`
	Normalized    bool          `json:"normalized,omitempty"`      // Specifies whether integer data values should be normalized.
	Count         uint32        `json:"count" validate:"required"` // The number of attributes referenced by this accessor.
	Type          AccessorType  `json:"type" validate:"lte=6"`
	Max           []float64     `json:"max,omitempty" validate:"omitempty,lte=16"` // Maximum value of each component in this attribute.
	Min           []float64     `json:"min,omitempty" validate:"omitempty,lte=16"` // Minimum value of each component in this attribute.
	Sparse        *Sparse       `json:"sparse,omitempty"`                          // Sparse storage of attributes that deviate from their initialization value.
}

// Sparse storage of attributes that deviate from their initialization value.
type Sparse struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Count      uint32        `json:"count" validate:"gte=1"` // Number of entries stored in the sparse array.
	Indices    SparseIndices `json:"indices"`                // Index array of size count that points to those accessor attributes that deviate from their initialization value.
	Values     SparseValues  `json:"values"`                 // Array of size count times number of components, storing the displaced accessor attributes pointed by indices.
}

// SparseValues stores the displaced accessor attributes pointed by accessor.sparse.indices.
type SparseValues struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	BufferView uint32      `json:"bufferView"`
	ByteOffset uint32      `json:"byteOffset,omitempty"`
}

// SparseIndices defines the indices of those attributes that deviate from their initialization value.
type SparseIndices struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	BufferView    uint32        `json:"bufferView"`
	ByteOffset    uint32        `json:"byteOffset,omitempty"`
	ComponentType ComponentType `json:"componentType" validate:"oneof=2 4 5"`
}

// A Buffer points to binary geometry, animation, or skins.
// If Data length is 0 and the Buffer is an external resource the Data won't be flushed,
// which can be useful when there is no need to load data in memory.
type Buffer struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	URI        string      `json:"uri,omitempty" validate:"omitempty"`
	ByteLength uint32      `json:"byteLength" validate:"required"`
	Data       []uint8     `json:"-"`
}

// IsEmbeddedResource returns true if the buffer points to an embedded resource.
func (b *Buffer) IsEmbeddedResource() bool {
	return strings.HasPrefix(b.URI, mimetypeApplicationOctet)
}

// EmbeddedResource defines the buffer as an embedded resource and encodes the URI so it points to the the resource.
func (b *Buffer) EmbeddedResource() {
	b.URI = fmt.Sprintf("%s,%s", mimetypeApplicationOctet, base64.StdEncoding.EncodeToString(b.Data))
}

// marshalData decode the buffer from the URI. If the buffer is not en embedded resource the returned array will be empty.
func (b *Buffer) marshalData() ([]uint8, error) {
	if !b.IsEmbeddedResource() {
		return nil, nil
	}
	startPos := len(mimetypeApplicationOctet) + 1
	sl, err := base64.StdEncoding.DecodeString(b.URI[startPos:])
	if len(sl) == 0 || err != nil {
		return nil, err
	}
	return sl, nil
}

// BufferView is a view into a buffer generally representing a subset of the buffer.
type BufferView struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Buffer     uint32      `json:"buffer"`
	ByteOffset uint32      `json:"byteOffset,omitempty"`
	ByteLength uint32      `json:"byteLength" validate:"required"`
	ByteStride uint32      `json:"byteStride,omitempty" validate:"omitempty,gte=4,lte=252"`
	Target     Target      `json:"target,omitempty" validate:"omitempty,oneof=34962 34963"`
}

// The Scene contains a list of root nodes.
type Scene struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Nodes      []uint32    `json:"nodes,omitempty" validate:"omitempty,unique"`
}

// A Node in the node hierarchy.
// It can have either a matrix or any combination of translation/rotation/scale (TRS) properties.
type Node struct {
	Extensions  Extensions  `json:"extensions,omitempty"`
	Extras      interface{} `json:"extras,omitempty"`
	Name        string      `json:"name,omitempty"`
	Camera      *uint32     `json:"camera,omitempty"`
	Children    []uint32    `json:"children,omitempty" validate:"omitempty,unique"`
	Skin        *uint32     `json:"skin,omitempty"`
	Matrix      [16]float64 `json:"matrix"` // A 4x4 transformation matrix stored in column-major order.
	Mesh        *uint32     `json:"mesh,omitempty"`
	Rotation    [4]float64  `json:"rotation" validate:"omitempty,dive,gte=-1,lte=1"` // The node's unit quaternion rotation in the order (x, y, z, w), where w is the scalar.
	Scale       [3]float64  `json:"scale"`
	Translation [3]float64  `json:"translation"`
	Weights     []float64   `json:"weights,omitempty"` // The weights of the instantiated Morph Target.
}

// MatrixOrDefault returns the node matrix if it represents a valid affine matrix, else return the default one.
func (n *Node) MatrixOrDefault() [16]float64 {
	if n.Matrix == emptyMatrix {
		return DefaultMatrix
	}
	return n.Matrix
}

// RotationOrDefault returns the node rotation if it represents a valid quaternion, else return the default one.
func (n *Node) RotationOrDefault() [4]float64 {
	if n.Rotation == emptyRotation {
		return DefaultRotation
	}
	return n.Rotation
}

// ScaleOrDefault returns the node scale if it represents a valid scale factor, else return the default one.
func (n *Node) ScaleOrDefault() [3]float64 {
	if n.Scale == emptyScale {
		return DefaultScale
	}
	return n.Scale
}

// TranslationOrDefault returns the node translation.
func (n *Node) TranslationOrDefault() [3]float64 {
	return n.Translation
}

// UnmarshalJSON unmarshal the node with the correct default values.
func (n *Node) UnmarshalJSON(data []byte) error {
	type alias Node
	tmp := alias(Node{
		Matrix:   DefaultMatrix,
		Rotation: DefaultRotation,
		Scale:    DefaultScale,
	})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = Node(tmp)
	}
	return err
}

// MarshalJSON marshal the node with the correct default values.
func (n *Node) MarshalJSON() ([]byte, error) {
	type alias Node
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(n)})
	if err == nil {
		if n.Matrix == DefaultMatrix {
			out = removeProperty([]byte(`"matrix":[1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1]`), out)
		} else if n.Matrix == emptyMatrix {
			out = removeProperty([]byte(`"matrix":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]`), out)
		}

		if n.Rotation == DefaultRotation {
			out = removeProperty([]byte(`"rotation":[0,0,0,1]`), out)
		} else if n.Rotation == emptyRotation {
			out = removeProperty([]byte(`"rotation":[0,0,0,0]`), out)
		}

		if n.Scale == DefaultScale {
			out = removeProperty([]byte(`"scale":[1,1,1]`), out)
		} else if n.Scale == emptyScale {
			out = removeProperty([]byte(`"scale":[0,0,0]`), out)
		}

		if n.Translation == DefaultTranslation {
			out = removeProperty([]byte(`"translation":[0,0,0]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// Skin defines joints and matrices.
type Skin struct {
	Extensions          Extensions  `json:"extensions,omitempty"`
	Extras              interface{} `json:"extras,omitempty"`
	Name                string      `json:"name,omitempty"`
	InverseBindMatrices *uint32     `json:"inverseBindMatrices,omitempty"`      // The index of the accessor containing the floating-point 4x4 inverse-bind matrices.
	Skeleton            *uint32     `json:"skeleton,omitempty"`                 // The index of the node used as a skeleton root. When undefined, joints transforms resolve to scene root.
	Joints              []uint32    `json:"joints" validate:"omitempty,unique"` // Indices of skeleton nodes, used as joints in this skin.
}

// A Camera projection. A node can reference a camera to apply a transform to place the camera in the scene.
type Camera struct {
	Extensions   Extensions    `json:"extensions,omitempty"`
	Extras       interface{}   `json:"extras,omitempty"`
	Name         string        `json:"name,omitempty"`
	Orthographic *Orthographic `json:"orthographic,omitempty"`
	Perspective  *Perspective  `json:"perspective,omitempty"`
}

// MarshalJSON marshal the camera with the correct default values.
func (c *Camera) MarshalJSON() ([]byte, error) {
	type alias Camera
	if c.Perspective != nil {
		return json.Marshal(&struct {
			Type string `json:"type"`
			*alias
		}{
			Type:  "perspective",
			alias: (*alias)(c),
		})
	} else if c.Orthographic != nil {
		return json.Marshal(&struct {
			Type string `json:"type"`
			*alias
		}{
			Type:  "orthographic",
			alias: (*alias)(c),
		})
	}
	return nil, errors.New("gltf: camera must defined either the perspective or orthographic property")
}

// Orthographic camera containing properties to create an orthographic projection matrix.
type Orthographic struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Xmag       float64     `json:"xmag"`                               // The horizontal magnification of the view.
	Ymag       float64     `json:"ymag"`                               // The vertical magnification of the view.
	Zfar       float64     `json:"zfar" validate:"gt=0,gtfield=Znear"` // The distance to the far clipping plane.
	Znear      float64     `json:"znear" validate:"gte=0"`             // The distance to the near clipping plane.
}

// Perspective camera containing properties to create a perspective projection matrix.
type Perspective struct {
	Extensions  Extensions  `json:"extensions,omitempty"`
	Extras      interface{} `json:"extras,omitempty"`
	AspectRatio *float64    `json:"aspectRatio,omitempty"`
	Yfov        float64     `json:"yfov"`           // The vertical field of view in radians.
	Zfar        *float64    `json:"zfar,omitempty"` // The distance to the far clipping plane.
	Znear       float64     `json:"znear"`          // The distance to the near clipping plane.
}

// A Mesh is a set of primitives to be rendered. A node can contain one mesh. A node's transform places the mesh in the scene.
type Mesh struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Primitives []Primitive `json:"primitives" validate:"required,gt=0,dive"`
	Weights    []float64   `json:"weights,omitempty"`
}

// Primitive defines the geometry to be rendered with the given material.
type Primitive struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Attributes Attribute     `json:"attributes"`
	Indices    *uint32       `json:"indices,omitempty"` // The index of the accessor that contains the indices.
	Material   *uint32       `json:"material,omitempty"`
	Mode       PrimitiveMode `json:"mode,omitempty" validate:"lte=6"`
	Targets    []Attribute   `json:"targets,omitempty" validate:"omitempty,dive,dive,keys,oneof=POSITION NORMAL TANGENT,endkeys"` // Only POSITION, NORMAL, and TANGENT supported.
}

// The Material appearance of a primitive.
type Material struct {
	Extensions           Extensions            `json:"extensions,omitempty"`
	Extras               interface{}           `json:"extras,omitempty"`
	Name                 string                `json:"name,omitempty"`
	PBRMetallicRoughness *PBRMetallicRoughness `json:"pbrMetallicRoughness,omitempty"`
	NormalTexture        *NormalTexture        `json:"normalTexture,omitempty"`
	OcclusionTexture     *OcclusionTexture     `json:"occlusionTexture,omitempty"`
	EmissiveTexture      *TextureInfo          `json:"emissiveTexture,omitempty"`
	EmissiveFactor       [3]float64            `json:"emissiveFactor,omitempty" validate:"dive,gte=0,lte=1"`
	AlphaMode            AlphaMode             `json:"alphaMode,omitempty" validate:"lte=2"`
	AlphaCutoff          *float64              `json:"alphaCutoff,omitempty" validate:"omitempty,gte=0"`
	DoubleSided          bool                  `json:"doubleSided,omitempty"`
}

// AlphaCutoffOrDefault returns the scale if it is not nil, else return the default one.
func (m *Material) AlphaCutoffOrDefault() float64 {
	if m.AlphaCutoff == nil {
		return 0.5
	}
	return *m.AlphaCutoff
}

// UnmarshalJSON unmarshal the material with the correct default values.
func (m *Material) UnmarshalJSON(data []byte) error {
	type alias Material
	tmp := alias(Material{AlphaCutoff: Float64(0.5)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*m = Material(tmp)
	}
	return err
}

// MarshalJSON marshal the material with the correct default values.
func (m *Material) MarshalJSON() ([]byte, error) {
	type alias Material
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(m)})
	if err == nil {
		if m.AlphaCutoff != nil && *m.AlphaCutoff == 0.5 {
			out = removeProperty([]byte(`"alphaCutoff":0.5`), out)
		}
		if m.EmissiveFactor == [3]float64{0, 0, 0} {
			out = removeProperty([]byte(`"emissiveFactor":[0,0,0]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// A NormalTexture references to a normal texture.
type NormalTexture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      *uint32     `json:"index,omitempty"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
	Scale      *float64    `json:"scale,omitempty"`
}

// ScaleOrDefault returns the scale if it is not nil, else return the default one.
func (n *NormalTexture) ScaleOrDefault() float64 {
	if n.Scale == nil {
		return 1
	}
	return *n.Scale
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (n *NormalTexture) UnmarshalJSON(data []byte) error {
	type alias NormalTexture
	tmp := alias(NormalTexture{Scale: Float64(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*n = NormalTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (n *NormalTexture) MarshalJSON() ([]byte, error) {
	type alias NormalTexture
	if n.Scale != nil && *n.Scale == 1 {
		return json.Marshal(&struct {
			Scale float64 `json:"scale,omitempty"`
			*alias
		}{
			Scale: 0,
			alias: (*alias)(n),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(n)})
}

// An OcclusionTexture references to an occlusion texture
type OcclusionTexture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      *uint32     `json:"index,omitempty"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
	Strength   *float64    `json:"strength,omitempty" validate:"omitempty,gte=0,lte=1"`
}

// StrengthOrDefault returns the strength if it is not nil, else return the default one.
func (o *OcclusionTexture) StrengthOrDefault() float64 {
	if o.Strength == nil {
		return 1
	}
	return *o.Strength
}

// UnmarshalJSON unmarshal the texture info with the correct default values.
func (o *OcclusionTexture) UnmarshalJSON(data []byte) error {
	type alias OcclusionTexture
	tmp := alias(OcclusionTexture{Strength: Float64(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*o = OcclusionTexture(tmp)
	}
	return err
}

// MarshalJSON marshal the texture info with the correct default values.
func (o *OcclusionTexture) MarshalJSON() ([]byte, error) {
	type alias OcclusionTexture
	if o.Strength != nil && *o.Strength == 1 {
		return json.Marshal(&struct {
			Strength float64 `json:"strength,omitempty"`
			*alias
		}{
			Strength: 0,
			alias:    (*alias)(o),
		})
	}
	return json.Marshal(&struct{ *alias }{alias: (*alias)(o)})
}

// The RGBA components of a color.
// Each element must be greater than or equal to 0 and less than or equal to 1.
type RGBA struct {
	R, G, B, A float64 `validate:"gte=0,lte=1"`
}

// NewRGBA returns a default RGBA color.
func NewRGBA() *RGBA {
	return &RGBA{1, 1, 1, 1}
}

// UnmarshalJSON unmarshal the color with the correct default values.
func (c *RGBA) UnmarshalJSON(data []byte) error {
	tmp := [4]float64{1, 1, 1, 1}
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		c.R, c.G, c.B, c.A = tmp[0], tmp[1], tmp[2], tmp[3]
	}
	return err
}

// MarshalJSON marshal the color with the correct default values.
func (c *RGBA) MarshalJSON() ([]byte, error) {
	return json.Marshal([4]float64{c.R, c.G, c.B, c.A})
}

// The RGB components of a color.
// Each element must be greater than or equal to 0 and less than or equal to 1.
type RGB struct {
	R, G, B float64 `validate:"gte=0,lte=1"`
}

// NewRGB returns a default RGB color.
func NewRGB() *RGB {
	return &RGB{1, 1, 1}
}

// UnmarshalJSON unmarshal the color with the correct default values.
func (c *RGB) UnmarshalJSON(data []byte) error {
	tmp := [3]float64{1, 1, 1}
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		c.R, c.G, c.B = tmp[0], tmp[1], tmp[2]
	}
	return err
}

// MarshalJSON marshal the color with the correct default values.
func (c *RGB) MarshalJSON() ([]byte, error) {
	return json.Marshal([3]float64{c.R, c.G, c.B})
}

// PBRMetallicRoughness defines a set of parameter values that are used to define the metallic-roughness material model from Physically-Based Rendering (PBR) methodology.
type PBRMetallicRoughness struct {
	Extensions               Extensions   `json:"extensions,omitempty"`
	Extras                   interface{}  `json:"extras,omitempty"`
	BaseColorFactor          *RGBA        `json:"baseColorFactor,omitempty"`
	BaseColorTexture         *TextureInfo `json:"baseColorTexture,omitempty"`
	MetallicFactor           *float64     `json:"metallicFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	RoughnessFactor          *float64     `json:"roughnessFactor,omitempty" validate:"omitempty,gte=0,lte=1"`
	MetallicRoughnessTexture *TextureInfo `json:"metallicRoughnessTexture,omitempty"`
}

// MetallicFactorOrDefault returns the metallic factor if it is not nil, else return the default one.
func (p *PBRMetallicRoughness) MetallicFactorOrDefault() float64 {
	if p.MetallicFactor == nil {
		return 1
	}
	return *p.MetallicFactor
}

// RoughnessFactorOrDefault returns the roughness factor if it is not nil, else return the default one.
func (p *PBRMetallicRoughness) RoughnessFactorOrDefault() float64 {
	if p.RoughnessFactor == nil {
		return 1
	}
	return *p.RoughnessFactor
}

// BaseColorFactorOrDefault returns the base color factor if it is not nil, else return the default one.
func (p *PBRMetallicRoughness) BaseColorFactorOrDefault() RGBA {
	if p.BaseColorFactor == nil {
		return *NewRGBA()
	}
	return *p.BaseColorFactor
}

// UnmarshalJSON unmarshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) UnmarshalJSON(data []byte) error {
	type alias PBRMetallicRoughness
	tmp := alias(PBRMetallicRoughness{BaseColorFactor: NewRGBA(), MetallicFactor: Float64(1), RoughnessFactor: Float64(1)})
	err := json.Unmarshal(data, &tmp)
	if err == nil {
		*p = PBRMetallicRoughness(tmp)
	}
	return err
}

// MarshalJSON marshal the pbr with the correct default values.
func (p *PBRMetallicRoughness) MarshalJSON() ([]byte, error) {
	type alias PBRMetallicRoughness
	out, err := json.Marshal(&struct{ *alias }{alias: (*alias)(p)})
	if err == nil {
		if p.MetallicFactor != nil && *p.MetallicFactor == 1 {
			out = removeProperty([]byte(`"metallicFactor":1`), out)
		}
		if p.RoughnessFactor != nil && *p.RoughnessFactor == 1 {
			out = removeProperty([]byte(`"roughnessFactor":1`), out)
		}
		if p.BaseColorFactor != nil && *p.BaseColorFactor == *NewRGBA() {
			out = removeProperty([]byte(`"baseColorFactor":[1,1,1,1]`), out)
		}
		out = sanitizeJSON(out)
	}
	return out, err
}

// TextureInfo references to a texture.
type TextureInfo struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Index      uint32      `json:"index"`
	TexCoord   uint32      `json:"texCoord,omitempty"` // The index of texture's TEXCOORD attribute used for texture coordinate mapping.
}

// A Texture and its sampler.
type Texture struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	Sampler    *uint32     `json:"sampler,omitempty"`
	Source     *uint32     `json:"source,omitempty"`
}

// Sampler of a texture for filtering and wrapping modes.
type Sampler struct {
	Extensions Extensions   `json:"extensions,omitempty"`
	Extras     interface{}  `json:"extras,omitempty"`
	Name       string       `json:"name,omitempty"`
	MagFilter  MagFilter    `json:"magFilter,omitempty" validate:"lte=1"`
	MinFilter  MinFilter    `json:"minFilter,omitempty" validate:"lte=5"`
	WrapS      WrappingMode `json:"wrapS,omitempty" validate:"lte=2"`
	WrapT      WrappingMode `json:"wrapT,omitempty" validate:"lte=2"`
}

// Image data used to create a texture. Image can be referenced by URI or bufferView index.
// mimeType is required in the latter case.
type Image struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Name       string      `json:"name,omitempty"`
	URI        string      `json:"uri,omitempty" validate:"omitempty"`
	MimeType   string      `json:"mimeType,omitempty" validate:"omitempty,oneof=image/jpeg image/png"` // Manadatory if BufferView is defined.
	BufferView uint32      `json:"bufferView,omitempty"`                                               // Use this instead of the image's uri property.
}

// IsEmbeddedResource returns true if the buffer points to an embedded resource.
func (im *Image) IsEmbeddedResource() bool {
	return strings.HasPrefix(im.URI, mimetypeImagePNG) || strings.HasPrefix(im.URI, mimetypeImageJPG)
}

// MarshalData decode the image from the URI. If the image is not en embedded resource the returned array will be empty.
func (im *Image) MarshalData() ([]uint8, error) {
	if !im.IsEmbeddedResource() {
		return []uint8{}, nil
	}
	mimetype := mimetypeImagePNG
	if strings.HasPrefix(im.URI, mimetypeImageJPG) {
		mimetype = mimetypeImageJPG
	}
	startPos := len(mimetype) + 1
	return base64.StdEncoding.DecodeString(im.URI[startPos:])
}

// An Animation keyframe.
type Animation struct {
	Extensions Extensions         `json:"extensions,omitempty"`
	Extras     interface{}        `json:"extras,omitempty"`
	Name       string             `json:"name,omitempty"`
	Channels   []Channel          `json:"channels" validate:"required,gt=0,dive"`
	Samplers   []AnimationSampler `json:"samplers" validate:"required,gt=0,dive"`
}

// AnimationSampler combines input and output accessors with an interpolation algorithm to define a keyframe graph (but not its target).
type AnimationSampler struct {
	Extensions    Extensions    `json:"extensions,omitempty"`
	Extras        interface{}   `json:"extras,omitempty"`
	Input         *uint32       `json:"input,omitempty"` // The index of an accessor containing keyframe input values.
	Interpolation Interpolation `json:"interpolation,omitempty" validate:"lte=2"`
	Output        *uint32       `json:"output,omitempty"` // The index of an accessor containing keyframe output values.
}

// The Channel targets an animation's sampler at a node's property.
type Channel struct {
	Extensions Extensions    `json:"extensions,omitempty"`
	Extras     interface{}   `json:"extras,omitempty"`
	Sampler    *uint32       `json:"sampler,omitempty"`
	Target     ChannelTarget `json:"target"`
}

// ChannelTarget describes the index of the node and TRS property that an animation channel targets.
// The Path represents the name of the node's TRS property to modify, or the "weights" of the Morph Targets it instantiates.
// For the "translation" property, the values that are provided by the sampler are the translation along the x, y, and z axes.
// For the "rotation" property, the values are a quaternion in the order (x, y, z, w), where w is the scalar.
// For the "scale" property, the values are the scaling factors along the x, y, and z axes.
type ChannelTarget struct {
	Extensions Extensions  `json:"extensions,omitempty"`
	Extras     interface{} `json:"extras,omitempty"`
	Node       *uint32     `json:"node,omitempty"`
	Path       TRSProperty `json:"path" validate:"lte=4"`
}

// Extensions is map where the keys are the extension identifiers and the values are the extensions payloads.
// If a key matches with one of the supported extensions the value will be marshalled as a pointer to the extension struct.
// If a key does not match with any of the supported extensions the value will be a json.RawMessage so its decoding can be delayed.
type Extensions map[string]interface{}

type envelope map[string]json.RawMessage

var extensions = make(map[string]func() json.Unmarshaler)

// RegisterExtension registers a function that returns a new extension of the given
// byte array. This is intended to be called from the init function in
// packages that implement extensions.
func RegisterExtension(key string, f func() json.Unmarshaler) {
	extensions[key] = f
}

// UnmarshalJSON unmarshal the extensions with the supported extensions initialized.
func (ext *Extensions) UnmarshalJSON(data []byte) error {
	if len(*ext) == 0 {
		*ext = make(Extensions)
	}
	var raw envelope
	err := json.Unmarshal(data, &raw)
	if err == nil {
		for key, value := range raw {
			if extFactory, ok := extensions[key]; ok {
				n := extFactory()
				err := json.Unmarshal(value, n)
				if err != nil {
					(*ext)[key] = value
				} else {
					(*ext)[key] = n
				}
			} else {
				(*ext)[key] = value
			}
		}
	}

	return err
}

func removeProperty(str []byte, b []byte) []byte {
	b = bytes.Replace(b, str, []byte(""), 1)
	return bytes.Replace(b, []byte(`,,`), []byte(","), 1)
}

func sanitizeJSON(b []byte) []byte {
	b = bytes.Replace(b, []byte(`{,`), []byte("{"), 1)
	return bytes.Replace(b, []byte(`,}`), []byte("}"), 1)
}
