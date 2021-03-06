package helium

import "errors"

const (
	XMLNamespace = "http://www.w3.org/XML/1998/namespace"
	XMLNsPrefix  = "xmlns"
	XMLPrefix    = "xml"
	XMLTextNoEnc = "textnoenc"
)

type LoadSubsetOption int

const (
	DetectIDs LoadSubsetOption = 1<<iota+1
	CompleteAttrs
	SkipIDs
)

type ParseOption int

// Note: Many of these are totally unimplemented at this point
const (
	ParseRecover   ParseOption = 1 << iota /* recover on errors */
	ParseNoEnt                             /* substitute entities */
	ParseDTDLoad                           /* load the external subset */
	ParseDTDAttr                           /* default DTD attributes */
	ParseDTDValid                          /* validate with the DTD */
	ParseNoError                           /* suppress error reports */
	ParseNoWarning                         /* suppress warning reports */
	ParsePedantic                          /* pedantic error reporting */
	ParseNoBlanks                          /* remove blank nodes */
	// gap here: ParseSAX1 is not implemented
	ParseXInclude   ParseOption = 1<<iota + 10 /* Implement XInclude substitition  */
	ParseNoNet                                 /* Forbid network access */
	ParseNoDict                                /* Do not reuse the context dictionnary */
	ParseNsClean                               /* remove redundant namespaces declarations */
	ParseNoCDATA                               /* merge CDATA as text nodes */
	ParseNoXIncNode                            /* do not generate XINCLUDE START/END nodes */
	ParseCompact                               /* compact small text nodes; no modification of the tree allowed afterwards (will possibly crash if you try to modify the tree) */
	// ParseOld10 is not implemented
	ParseNoBaseFix /* do not fixup XINCLUDE xml:base uris */
	ParseHuge      /* relax any hardcoded limit from the parser */
	// ParseOldSAX is not implemented
	ParseIgnoreEnc ParseOption = 1<<iota + 21 /* ignore internal document encoding hint */
	ParseBigLines  ParseOption = 1 << 22      /* Store big lines numbers in text PSVI field */
)

type AttributeType int

const (
	AttrInvalid AttributeType = iota
	AttrCDATA
	AttrID
	AttrIDRef
	AttrIDRefs
	AttrEntity
	AttrEntities
	AttrNmtoken
	AttrNmtokens
	AttrEnumeration
	AttrNotation
)

type AttributeDefault int

const (
	AttrDefaultInvalid AttributeDefault = iota
	AttrDefaultNone
	AttrDefaultRequired
	AttrDefaultImplied
	AttrDefaultFixed
)

type Enumeration []string

var (
	ErrNilNode            = errors.New("nil node")
	ErrInvalidOperation   = errors.New("operation cannot be performed")
	ErrDuplicateAttribute = errors.New("duplicate attribute")
)

type ErrUnimplemented struct {
	target string
}

type Node interface {
	setLastChild(Node)
	setFirstChild(Node)

	AddChild(Node) error
	AddContent([]byte) error
	AddSibling(Node) error
	Content() []byte
	FirstChild() Node
	LastChild() Node
	Name() string
	NextSibling() Node
	OwnerDocument() *Document
	Parent() Node
	PrevSibling() Node
	Replace(Node)
	SetNextSibling(Node)
	SetOwnerDocument(doc *Document)
	SetParent(Node)
	SetPrevSibling(Node)
	SetTreeDoc(doc *Document)
	Type() ElementType
}

// docnode is responsible for handling the basic tree-ish operations
type docnode struct {
	name       string
	etype      ElementType
	firstChild Node
	lastChild  Node
	parent     Node
	next       Node
	prev       Node
	doc        *Document
}

// node represents a node in a XML tree.
type node struct {
	docnode
	private    interface{}
	content    []byte
	properties *Attribute
	ns         *Namespace
	nsDefs     []*Namespace
}

type DocumentStandaloneType int

const (
	StandaloneInvalidValue = -99
	StandaloneExplicitYes  = 1
	StandaloneExplicitNo   = 0
	StandaloneNoXMLDecl    = -1
	StandaloneImplicitNo   = -2
)

type Document struct {
	docnode
	version    string
	encoding   string
	standalone DocumentStandaloneType

	intSubset *DTD
	extSubset *DTD
}

type ProcessingInstruction struct {
	docnode
	target string
	data   string
}

type DTD struct {
	docnode
	attributes map[string]*AttributeDecl
	elements   map[string]*ElementDecl
	entities   map[string]*Entity
	pentities  map[string]*Entity
	externalID string
	systemID   string
}

type Namespace struct {
	etype   ElementType
	href    string
	prefix  string
	context *Document
}

type Attribute struct {
	docnode
	atype       AttributeType
	defaultAttr bool
	ns          *Namespace
}

type ElementType int

const (
	ElementNode ElementType = iota + 1
	AttributeNode
	TextNode
	CDATASectionNode
	EntityRefNode
	EntityNode
	ProcessingInstructionNode
	CommentNode
	DocumentNode
	DocumentTypeNode
	DocumentFragNode
	NotationNode
	HTMLDocumentNode
	DTDNode
	ElementDeclNode
	AttributeDeclNode
	EntityDeclNode
	NamespaceDeclNode
	XIncludeStartNode
	XIncludeEndNode

	// This doesn't exist in libxml2. Do we need it?
	NamespaceNode
)

type NamespaceContainer interface {
	Namespaces() []*Namespace
}

type EntityRef struct {
	node
}

// Text is just a wrapper around Node so that we can
// use Go-ish type checks
type Text struct {
	node
}

// Comment is just a wrapper around Node so that we can
// use Go-ish type checks
type Comment struct {
	node
}

// Element is just a wrapper around Node so that we can
// use Go-ish type checks
type Element struct {
	node
}

// Nemaspacer is an interface for things that has a namespace
// prefix and uri
type Namespacer interface {
	Namespace() *Namespace
	Namespaces() []*Namespace
	Prefix() string
	URI() string
	LocalName() string
}

// AttributeDecl is an xml attribute delcaration from DTD
type AttributeDecl struct {
	docnode
	atype    AttributeType    // attribute type
	def      AttributeDefault // default
	defvalue string           // ... or the default value
	tree     Enumeration      // ... or the numeration tree, if any
	prefix   string           // the namespace prefix, if any
	elem     string           // name of the element holding the attribute
}

// ElementDecl is an xml element declaration from DTD
type ElementDecl struct {
	docnode
	decltype   ElementTypeVal
	content    *ElementContent
	attributes *AttributeDecl
	prefix     string
	// xmlRegexpPtr contModel
}

// ElementTypeVal represents the different possibilities for an element
// content type.
type ElementTypeVal int

const (
	UndefinedElementType ElementTypeVal = iota
	EmptyElementType
	AnyElementType
	MixedElementType
	ElementElementType
)

type ElementContentType int

const (
	ElementContentPCDATA ElementContentType = iota + 1
	ElementContentElement
	ElementContentSeq
	ElementContentOr
)

type ElementContentOccur int

const (
	ElementContentOnce ElementContentOccur = iota + 1
	ElementContentOpt
	ElementContentMult
	ElementContentPlus
)

type ElementContent struct {
	// XXX no doc?
	ctype  ElementContentType
	coccur ElementContentOccur
	name   string
	prefix string
	c1     *ElementContent
	c2     *ElementContent
	parent *ElementContent
}

type EntityType int

const (
	InternalGeneralEntity EntityType = iota + 1
	ExternalGeneralParsedEntity
	ExternalGeneralUnparsedEntity
	InternalParameterEntity
	ExternalParameterEntity
	InternalPredefinedEntity
)

type Entity struct {
	node
	orig       string     // content without substitution
	content    string     // content or ndata if unparsed
	entityType EntityType // the entity type
	externalID string     // external identifier for PUBLIC
	systemID   string     // URI for a SYSTEM or PUBLIC entity
	uri        string     // the full URI as computed
	owner      bool       // does the entity own children
	checked    int        // was the entity content checked
	/* this is also used to count entities
	 * references done from that entity
	 * and if it contains '<' */
}

var (
	EntityLT         = newEntity("lt", InternalPredefinedEntity, "", "", "<", "&lt;")
	EntityGT         = newEntity("gt", InternalPredefinedEntity, "", "", ">", "&gt;")
	EntityAmpersand  = newEntity("amp", InternalPredefinedEntity, "", "", "&", "&amp;")
	EntityApostrophe = newEntity("apos", InternalPredefinedEntity, "", "", "'", "&apos;")
	EntityQuote      = newEntity("quot", InternalPredefinedEntity, "", "", `"`, "&quot;")
)
