package helium

import (
	"errors"

	"github.com/lestrrat/helium/internal/debug"
	"github.com/lestrrat/helium/sax"
)

type TreeBuilder struct {
	doc  *Document
	node Node
}

func (t *TreeBuilder) SetDocumentLocator(ctxif sax.Context, loc sax.DocumentLocator) error {
	return nil
}

func (t *TreeBuilder) StartDocument(ctxif sax.Context) error {
	if debug.Enabled {
		g := debug.IPrintf("START tree.StartDocument")
		defer g.IRelease("END tree.StartDocument")
	}

	ctx := ctxif.(*parserCtx)

	t.doc = NewDocument(ctx.version, ctx.encoding, ctx.standalone)
	return nil
}

func (t *TreeBuilder) EndDocument(ctxif sax.Context) error {
	if debug.Enabled {
		g := debug.IPrintf("START tree.EndDocument")
		defer g.IRelease("END tree.EndDocument")
	}
	ctx := ctxif.(*parserCtx)
	ctx.doc = t.doc
	t.doc = nil
	return nil
}

func (t *TreeBuilder) ProcessingInstruction(ctxif sax.Context, target, data string) error {
	//	ctx := ctxif.(*parserCtx)
	pi, err := t.doc.CreatePI(target, data)
	if err != nil {
		return err
	}

	// register to the document
	t.doc.IntSubset().AddChild(pi)
	if t.node == nil {
		t.doc.AddChild(pi)
		return nil
	}

	// what's the "current" node?
	if t.node.Type() == ElementNode {
		t.node.AddChild(pi)
	} else {
//		t.node.AddSibling(pi)
		panic("unimplemented")
	}
	return nil
}

func (t *TreeBuilder) StartElement(ctxif sax.Context, localname, prefix, uri string, namespaces []sax.Namespace, attrs []sax.Attribute) error {
	//	ctx := ctxif.(*parserCtx)
	if debug.Enabled {
		if prefix != "" {
			debug.Printf("tree.StartElement: %s:%s", prefix, localname)
		} else {
			debug.Printf("tree.StartElement: %s", localname)
		}
	}
	e, err := t.doc.CreateElement(localname)
	if err != nil {
		return err
	}

	for _, attr := range attrs {
		e.SetAttribute(attr.Name(), attr.Value())
	}

	if t.node == nil {
		t.doc.AddChild(e)
	} else {
		t.node.AddChild(e)
	}

	t.node = e

	return nil
}

func (t *TreeBuilder) EndElement(ctxif sax.Context, localname, prefix, uri string) error {
	if debug.Enabled {
		if prefix != "" {
			debug.Printf("tree.EndElement: %s:%s", prefix, localname)
		} else {
			debug.Printf("tree.EndElement: %s", localname)
		}
	}
	ctx := ctxif.(*parserCtx)
	debug.Printf("t.node (%p) -----> t.node = ctx.peekNode (%p)",t.node, ctx.peekNode())
	if e, ok := t.node.(*Element); ok && e.LocalName() == localname && e.Prefix() == prefix && e.URI() == uri {
		t.node = t.node.Parent()
	}
	return nil
}

func (t *TreeBuilder) Characters(ctxif sax.Context, data []byte) error {
	if debug.Enabled {
		g := debug.IPrintf("START tree.Characters: '%s' (%v)", data, data)
		defer g.IRelease("END tree.Characters")
	}

	if t.node == nil {
		return errors.New("text content placed in wrong location")
	}

	return t.node.AddContent(data)
}

func (t *TreeBuilder) StartCDATA(_ sax.Context) error {
	return nil
}

func (t *TreeBuilder) EndCDATA(_ sax.Context) error {
	return nil
}

func (t *TreeBuilder) Comment(ctxif sax.Context, data []byte) error {
	if debug.Enabled {
		g := debug.IPrintf("START tree.Comment: %s", data)
		defer g.IRelease("END tree.Comment")
	}

	if t.node == nil {
		return errors.New("comment placed in wrong location")
	}

	e, err := t.doc.CreateComment(data)
	if err != nil {
		return err
	}
	t.node.AddChild(e)
	return nil
}

func (t *TreeBuilder) InternalSubset(ctxif sax.Context, name, eid, uri string) error {
	return nil
}

func (t *TreeBuilder) ExternalSubset(ctxif sax.Context, name, eid, uri string) error {
	return nil
}

func (t *TreeBuilder) GetEntity(ctxif sax.Context, name string) (*Entity, error) {
	ctx := ctxif.(*parserCtx)

	if ctx.inSubset == 0 {
		ret := resolvePredefinedEntity(name)
		if ret != nil {
			return ret, nil
		}
	}

	var ret *Entity
	var ok bool
	if ctx.doc == nil || ctx.doc.standalone != 1 {
		ret, _ = ctx.doc.GetEntity(name)
	} else {
		if ctx.inSubset == 2 {
			ctx.doc.standalone = 0
			ret, _ = ctx.doc.GetEntity(name)
			ctx.doc.standalone = 1
		} else {
			ret, ok = ctx.doc.GetEntity(name)
			if !ok {
				ctx.doc.standalone = 0
				ret, ok = ctx.doc.GetEntity(name)
				if !ok {
					return nil, errors.New("Entity(" + name + ") document marked standalone but requires eternal subset")
				}
				ctx.doc.standalone = 1
			}
		}
	}
/*
    if ((ret != NULL) &&
        ((ctxt->validate) || (ctxt->replaceEntities)) &&
        (ret->children == NULL) &&
        (ret->etype == XML_EXTERNAL_GENERAL_PARSED_ENTITY)) {
        int val;

        // for validation purposes we really need to fetch and
        // parse the external entity
        xmlNodePtr children;
        unsigned long oldnbent = ctxt->nbentities;

        val = xmlParseCtxtExternalEntity(ctxt, ret->URI,
                                         ret->ExternalID, &children);
        if (val == 0) {
            xmlAddChildList((xmlNodePtr) ret, children);
        } else {
            xmlFatalErrMsg(ctxt, XML_ERR_ENTITY_PROCESSING,
                           "Failure to process entity %s\n", name, NULL);
            ctxt->validate = 0;
            return(NULL);
        }
        ret->owner = 1;
        if (ret->checked == 0) {
            ret->checked = (ctxt->nbentities - oldnbent + 1) * 2;
            if ((ret->content != NULL) && (xmlStrchr(ret->content, '<')))
                ret->checked |= 1;
        }
    }
*/
	return ret, nil
}

func (t *TreeBuilder) GetParameterEntity(ctxif sax.Context, name string) (sax.Entity, error) {
	if ctxif == nil {
		return nil, ErrInvalidParserCtx
	}

	ctx := ctxif.(*parserCtx)
	doc := ctx.doc
	if doc == nil {
		return nil, ErrInvalidDocument
	}

	if ret, ok := doc.GetParameterEntity(name); ok {
		return ret, nil
	}

	return nil, ErrEntityNotFound
}

func (t *TreeBuilder) AttributeDecl(ctx sax.Context, eName string, aName string, typ int, deftype int, value sax.AttributeDefaultValue, enum sax.Enumeration) error {
	return nil
}

func (t *TreeBuilder) ElementDecl(ctx sax.Context, name string, typ int, content sax.ElementContent) error {
	return nil
}

func (t *TreeBuilder) EndDTD(ctx sax.Context) error {
	return nil
}

func (t *TreeBuilder) EndEntity(ctx sax.Context, name string) error {
	return nil
}
func (t *TreeBuilder) ExternalEntityDecl(ctx sax.Context, name string, publicID string, systemID string) error {
	return nil
}

func (t *TreeBuilder) GetExternalSubset(ctx sax.Context, name string, baseURI string) error {
	return nil
}

func (t *TreeBuilder) IgnorableWhitespace(ctxif sax.Context, content []byte) error {
	if debug.Enabled {
		g := debug.IPrintf("START tree.IgnorableWhitespace (%v)", content)
		defer g.IRelease("END tree.IgnorableWhitespace")
	}

	ctx := ctxif.(*parserCtx)
	if ctx.keepBlanks {
		return t.Characters(ctx, content)
	}
	return nil
}

func (t *TreeBuilder) InternalEntityDecl(ctx sax.Context, name string, value string) error {
	return nil
}

func (t *TreeBuilder) NotationDecl(ctx sax.Context, name string, publicID string, systemID string) error {
	return nil
}

func (t *TreeBuilder) Reference(ctx sax.Context, name string) error {
	return nil
}

func (t *TreeBuilder) ResolveEntity(ctx sax.Context, name string, publicID string, baseURI string, systemID string) (sax.Entity, error) {
	return nil, errors.New("entity not found")
}

func (t *TreeBuilder) SkippedEntity(ctx sax.Context, name string) error {
	return nil
}

func (t *TreeBuilder) StartDTD(ctx sax.Context, name string, publicID string, systemID string) error {
	return nil
}

func (t *TreeBuilder) StartEntity(ctx sax.Context, name string) error {
	return nil
}

func (t *TreeBuilder) UnparsedEntityDecl(ctx sax.Context, name string, typ int, publicID string, systemID string, notation string) error {
	return nil
}

