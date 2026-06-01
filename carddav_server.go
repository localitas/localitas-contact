package contact

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav/carddav"
	"gopkg.in/yaml.v3"
)

type CardDAVBackend struct {
	store *Store
}

func NewCardDAVBackend(store *Store) *CardDAVBackend {
	return &CardDAVBackend{store: store}
}

func NewCardDAVHandler(store *Store, prefix string) *carddav.Handler {
	return &carddav.Handler{
		Backend: NewCardDAVBackend(store),
		Prefix:  prefix,
	}
}

func (b *CardDAVBackend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/carddav/", nil
}

func (b *CardDAVBackend) AddressBookHomeSetPath(ctx context.Context) (string, error) {
	return "/carddav/", nil
}

func (b *CardDAVBackend) ListAddressBooks(ctx context.Context) ([]carddav.AddressBook, error) {
	return []carddav.AddressBook{{
		Path:        "/carddav/default/",
		Name:        "Contacts",
		Description: "Localitas Contacts",
	}}, nil
}

func (b *CardDAVBackend) GetAddressBook(ctx context.Context, path string) (*carddav.AddressBook, error) {
	return &carddav.AddressBook{
		Path:        "/carddav/default/",
		Name:        "Contacts",
		Description: "Localitas Contacts",
	}, nil
}

func (b *CardDAVBackend) CreateAddressBook(ctx context.Context, addressBook *carddav.AddressBook) error {
	return nil
}

func (b *CardDAVBackend) DeleteAddressBook(ctx context.Context, path string) error {
	return fmt.Errorf("cannot delete the default address book")
}

func (b *CardDAVBackend) GetAddressObject(ctx context.Context, path string, req *carddav.AddressDataRequest) (*carddav.AddressObject, error) {
	id := extractContactID(path)
	if id == "" {
		return nil, fmt.Errorf("contact not found")
	}
	c, err := b.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return contactToAddressObject(c), nil
}

func (b *CardDAVBackend) ListAddressObjects(ctx context.Context, path string, req *carddav.AddressDataRequest) ([]carddav.AddressObject, error) {
	contacts, err := b.store.List(ctx, "", 10000, 0)
	if err != nil {
		return nil, err
	}
	objects := make([]carddav.AddressObject, 0, len(contacts))
	for _, c := range contacts {
		objects = append(objects, *contactToAddressObject(c))
	}
	return objects, nil
}

func (b *CardDAVBackend) QueryAddressObjects(ctx context.Context, path string, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	return b.ListAddressObjects(ctx, path, &query.DataRequest)
}

func (b *CardDAVBackend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (*carddav.AddressObject, error) {
	data := vcardToContactData(card)
	id := extractContactID(path)

	if id != "" {
		existing, err := b.store.Get(ctx, id)
		if err == nil && existing != nil {
			b.store.Update(ctx, id, data)
			updated, _ := b.store.Get(ctx, id)
			return contactToAddressObject(updated), nil
		}
	}

	c, err := b.store.Create(ctx, "", data)
	if err != nil {
		return nil, err
	}
	return contactToAddressObject(c), nil
}

func (b *CardDAVBackend) DeleteAddressObject(ctx context.Context, path string) error {
	id := extractContactID(path)
	if id == "" {
		return fmt.Errorf("contact not found")
	}
	return b.store.Delete(ctx, id)
}

func extractContactID(path string) string {
	path = strings.TrimPrefix(path, "/carddav/default/")
	path = strings.TrimPrefix(path, "/carddav/")
	path = strings.TrimSuffix(path, ".vcf")
	path = strings.TrimSuffix(path, "/")
	return path
}

func contactToVCard(c *Contact) vcard.Card {
	var data ContactData
	yaml.Unmarshal([]byte(c.Data), &data)

	card := make(vcard.Card)
	card.SetValue(vcard.FieldVersion, "3.0")
	card.SetValue(vcard.FieldFormattedName, data.FullName)
	card.SetValue(vcard.FieldUID, c.ID+"@localitas")

	for _, email := range data.Emails {
		if email != "" {
			card.Add(vcard.FieldEmail, &vcard.Field{Value: email, Params: vcard.Params{"TYPE": {"INTERNET"}}})
		}
	}

	for _, phone := range data.Phones {
		if phone != "" {
			card.Add(vcard.FieldTelephone, &vcard.Field{Value: phone})
		}
	}

	if data.Work.Current != nil {
		if data.Work.Current.Company != "" {
			card.SetValue(vcard.FieldOrganization, data.Work.Current.Company)
		}
		if data.Work.Current.Title != "" {
			card.SetValue(vcard.FieldTitle, data.Work.Current.Title)
		}
	}

	if data.Social.Website != "" {
		card.Add(vcard.FieldURL, &vcard.Field{Value: data.Social.Website})
	}
	if data.Social.LinkedIn != "" {
		card.Add(vcard.FieldURL, &vcard.Field{Value: data.Social.LinkedIn, Params: vcard.Params{"TYPE": {"LinkedIn"}}})
	}

	card.SetValue(vcard.FieldRevision, c.UpdatedAt.UTC().Format("20060102T150405Z"))

	return card
}

func vcardToContactData(card vcard.Card) ContactData {
	var data ContactData

	if fn := card.PreferredValue(vcard.FieldFormattedName); fn != "" {
		data.FullName = fn
	}

	for _, f := range card[vcard.FieldEmail] {
		if f.Value != "" {
			data.Emails = append(data.Emails, f.Value)
		}
	}

	for _, f := range card[vcard.FieldTelephone] {
		if f.Value != "" {
			data.Phones = append(data.Phones, f.Value)
		}
	}

	if org := card.PreferredValue(vcard.FieldOrganization); org != "" {
		if data.Work.Current == nil {
			data.Work.Current = &Job{}
		}
		data.Work.Current.Company = org
	}

	if title := card.PreferredValue(vcard.FieldTitle); title != "" {
		if data.Work.Current == nil {
			data.Work.Current = &Job{}
		}
		data.Work.Current.Title = title
	}

	for _, f := range card[vcard.FieldURL] {
		if f.Value != "" {
			types := f.Params.Get("TYPE")
			if strings.EqualFold(types, "linkedin") {
				data.Social.LinkedIn = f.Value
			} else if data.Social.Website == "" {
				data.Social.Website = f.Value
			}
		}
	}

	return data
}

func contactToAddressObject(c *Contact) *carddav.AddressObject {
	card := contactToVCard(c)
	etag := fmt.Sprintf("%x", sha256.Sum256([]byte(c.ID+c.UpdatedAt.String())))
	return &carddav.AddressObject{
		Path:    "/carddav/default/" + c.ID + ".vcf",
		ModTime: c.UpdatedAt,
		ETag:    `"` + etag[:16] + `"`,
		Card:    card,
	}
}

func init() {
	_ = time.Now
}
