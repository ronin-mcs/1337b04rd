package domain

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"1337b04rd/models"
)

type fakePostsRepo struct {
	posts     []models.Post
	byID      map[int]*models.Post
	created   *models.Post
	withOP    *models.Anon
	updatedID int
}

func (f *fakePostsRepo) Create(post *models.Post) error {
	f.created = post
	if post.PostID == 0 {
		post.PostID = 10
	}
	return nil
}

func (f *fakePostsRepo) CreateWithOP(post *models.Post, anon *models.Anon) error {
	f.created = post
	f.withOP = anon
	if post.PostID == 0 {
		post.PostID = 10
	}
	if anon.AnonID == 0 {
		anon.AnonID = 20
	}
	return nil
}

func (f *fakePostsRepo) GetByID(id int, isActual bool) (*models.Post, error) {
	if f.byID == nil {
		return nil, nil
	}
	return f.byID[id], nil
}

func (f *fakePostsRepo) GetAll(isActual bool) ([]models.Post, error) { return f.posts, nil }
func (f *fakePostsRepo) UpdateStatus(id int) error                   { f.updatedID = id; return nil }
func (f *fakePostsRepo) Delete(id int) error                         { return nil }
func (f *fakePostsRepo) AssignOP(postID, anonID int) error           { return nil }

type fakeCommentsRepo struct {
	comments []models.Comment
	byParent map[int][]models.Comment
	anons    map[int]models.Anon
	created  *models.Comment
}

func (f *fakeCommentsRepo) Create(comment *models.Comment) error {
	f.created = comment
	if comment.CommentID == 0 {
		comment.CommentID = 30
	}
	return nil
}

func (f *fakeCommentsRepo) GetByID(id int) (*models.Comment, error) { return nil, nil }
func (f *fakeCommentsRepo) GetByPostID(postID int) ([]models.Comment, map[int][]models.Comment, map[int]models.Anon, error) {
	return f.comments, f.byParent, f.anons, nil
}
func (f *fakeCommentsRepo) GetAll() ([]models.Comment, error) { return f.comments, nil }
func (f *fakeCommentsRepo) Delete(id int) error               { return nil }

type fakeAnonsRepo struct {
	byID        map[int]*models.Anon
	avatarCount map[string]int
	created     *models.Anon
}

func (f *fakeAnonsRepo) Create(anon *models.Anon) error {
	f.created = anon
	if anon.AnonID == 0 {
		anon.AnonID = 40
	}
	return nil
}

func (f *fakeAnonsRepo) GetByID(id int) (*models.Anon, error) {
	if f.byID == nil || f.byID[id] == nil {
		return nil, errors.New("anon not found")
	}
	return f.byID[id], nil
}

func (f *fakeAnonsRepo) GetAll() ([]models.Anon, error) { return nil, nil }
func (f *fakeAnonsRepo) GetAllByPostID(id int) ([]models.Anon, error) {
	return nil, nil
}

func (f *fakeAnonsRepo) GetAvatarCountByPostID(id int) (map[string]int, error) {
	return f.avatarCount, nil
}
func (f *fakeAnonsRepo) Delete(id int) error             { return nil }
func (f *fakeAnonsRepo) DeleteByPostID(postID int) error { return nil }

type fakeSessionsRepo struct {
	byID    map[int]*models.Session
	created *models.Session
	updated *models.Session
}

func (f *fakeSessionsRepo) Create(session *models.Session) error {
	f.created = session
	if session.SessionID == 0 {
		session.SessionID = 50
	}
	if f.byID == nil {
		f.byID = map[int]*models.Session{}
	}
	f.byID[session.SessionID] = session
	return nil
}

func (f *fakeSessionsRepo) UpdateSessionHistory(session *models.Session) error {
	f.updated = session
	return nil
}

func (f *fakeSessionsRepo) GetByID(id int) (*models.Session, error) {
	if f.byID == nil {
		return nil, nil
	}
	return f.byID[id], nil
}

func (f *fakeSessionsRepo) GetAll() ([]models.Session, error) { return nil, nil }
func (f *fakeSessionsRepo) DeleteExpired() error              { return nil }
func (f *fakeSessionsRepo) Delete(id int) error               { return nil }

type fakeAvatarStorage struct {
	random int
	all    []int
}

func (f fakeAvatarStorage) GetRandomCharacterID() (int, error) { return f.random, nil }
func (f fakeAvatarStorage) GetAllCharacterIDs() []int          { return f.all }
func (f fakeAvatarStorage) GetAvatar(characterID int) (io.ReadCloser, string, error) {
	return io.NopCloser(strings.NewReader("")), "image/png", nil
}

type savedFile struct {
	key         string
	contentType string
	body        string
}

type fakeFileStorage struct {
	links map[string]string
	saved []savedFile
}

func (f *fakeFileStorage) SaveFile(fileKey string, fileData io.Reader, contentType string) error {
	body, err := io.ReadAll(fileData)
	if err != nil {
		return err
	}
	f.saved = append(f.saved, savedFile{key: fileKey, contentType: contentType, body: string(body)})
	return nil
}

func (f *fakeFileStorage) GetFileLink(fileKey string) (string, error) {
	if f.links != nil && f.links[fileKey] != "" {
		return f.links[fileKey], nil
	}
	return "https://files.test/" + fileKey, nil
}

func (f *fakeFileStorage) DeleteFile(fileKey string) error { return nil }

type fakeAttachmentsRepo struct {
	byPost    map[int][]models.Attachment
	byComment map[int][]models.Attachment
	created   []models.Attachment
}

func (f *fakeAttachmentsRepo) Create(attachment *models.Attachment) error {
	f.created = append(f.created, *attachment)
	return nil
}

func (f *fakeAttachmentsRepo) GetByPostID(postID int) ([]models.Attachment, error) {
	return f.byPost[postID], nil
}

func (f *fakeAttachmentsRepo) GetByCommentID(commentID int) ([]models.Attachment, error) {
	return f.byComment[commentID], nil
}

func (f *fakeAttachmentsRepo) DeleteByFileKey(fileKey string) error { return nil }

func newTestService(posts *fakePostsRepo, comments *fakeCommentsRepo, anons *fakeAnonsRepo, sessions *fakeSessionsRepo, files *fakeFileStorage, attachments *fakeAttachmentsRepo) *PostService {
	return NewPostService(
		fakeAvatarStorage{random: 7, all: []int{1, 2, 3}},
		files,
		posts,
		comments,
		anons,
		sessions,
		attachments,
	)
}

func TestCreateNewSessionID(t *testing.T) {
	sessions := &fakeSessionsRepo{}
	service := newTestService(&fakePostsRepo{}, &fakeCommentsRepo{}, &fakeAnonsRepo{}, sessions, &fakeFileStorage{}, &fakeAttachmentsRepo{})

	id, err := service.CreateNewSessionID()
	if err != nil {
		t.Fatalf("CreateNewSessionID() error = %v", err)
	}
	if id != 50 {
		t.Fatalf("CreateNewSessionID() = %d, want 50", id)
	}
	if sessions.created.Sessions == nil {
		t.Fatal("created session map is nil")
	}
}

func TestUploadSessionIDCreatesAndUpdatesSession(t *testing.T) {
	sessions := &fakeSessionsRepo{}
	service := newTestService(&fakePostsRepo{}, &fakeCommentsRepo{}, &fakeAnonsRepo{}, sessions, &fakeFileStorage{}, &fakeAttachmentsRepo{})

	if err := service.UploadSessionID(12, 99, 77); err != nil {
		t.Fatalf("UploadSessionID() error = %v", err)
	}
	if sessions.created == nil {
		t.Fatal("expected missing session to be created")
	}
	if got := sessions.updated.Sessions[12]; got != 77 {
		t.Fatalf("updated session post mapping = %d, want 77", got)
	}
}

func TestGetPostByID(t *testing.T) {
	post := &models.Post{PostID: 3, Title: "hello"}
	service := newTestService(&fakePostsRepo{byID: map[int]*models.Post{3: post}}, &fakeCommentsRepo{}, &fakeAnonsRepo{}, &fakeSessionsRepo{}, &fakeFileStorage{}, &fakeAttachmentsRepo{})

	got, err := service.GetActualPostByID(3)
	if err != nil {
		t.Fatalf("GetActualPostByID() error = %v", err)
	}
	if got.PostID != 3 {
		t.Fatalf("PostID = %d, want 3", got.PostID)
	}

	if _, err := service.GetActualPostByID(404); !errors.Is(err, models.ErrPostIsArchived) {
		t.Fatalf("missing actual post error = %v, want %v", err, models.ErrPostIsArchived)
	}
	if _, err := service.GetArchivedPostByID(404); !errors.Is(err, models.ErrPostIsNotArchived) {
		t.Fatalf("missing archived post error = %v, want %v", err, models.ErrPostIsNotArchived)
	}
}

func TestConstructCatalogPostViews(t *testing.T) {
	posts := []models.Post{{PostID: 1, AnonID: 10, Title: "topic"}}
	anons := &fakeAnonsRepo{byID: map[int]*models.Anon{10: {AnonID: 10, AnonName: "OP"}}}
	attachments := &fakeAttachmentsRepo{byPost: map[int][]models.Attachment{
		1: {{PostID: 1, FileKey: "posts/1/a.png", OriginalName: "a.png"}},
	}}
	files := &fakeFileStorage{}
	service := newTestService(&fakePostsRepo{}, &fakeCommentsRepo{}, anons, &fakeSessionsRepo{}, files, attachments)

	views, err := service.ConstructCatalogPostViews(posts)
	if err != nil {
		t.Fatalf("ConstructCatalogPostViews() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("views length = %d, want 1", len(views))
	}
	if views[0].Preview.Link == "" {
		t.Fatal("preview link is empty")
	}
	if views[0].Op.AnonName != "OP" {
		t.Fatalf("op name = %q, want OP", views[0].Op.AnonName)
	}
}

func TestConstructPostPagePostView(t *testing.T) {
	commentID := 101
	post := models.Post{PostID: 2, AnonID: 20, Title: "topic", CreatedAt: time.Now()}
	anons := &fakeAnonsRepo{byID: map[int]*models.Anon{20: {AnonID: 20, AnonName: "OP"}}}
	comments := &fakeCommentsRepo{
		byParent: map[int][]models.Comment{
			0: {{CommentID: commentID, PostID: 2, TextContent: "root", AnonID: 21}},
		},
		anons: map[int]models.Anon{
			commentID: {AnonID: 21, AnonName: "Anon21", Avatar: "5"},
		},
	}
	attachments := &fakeAttachmentsRepo{
		byPost: map[int][]models.Attachment{
			2: {{PostID: 2, FileKey: "posts/2/a.png"}},
		},
		byComment: map[int][]models.Attachment{
			commentID: {{PostID: 2, CommentID: &commentID, FileKey: "comments/2/101/b.png"}},
		},
	}
	service := newTestService(&fakePostsRepo{}, comments, anons, &fakeSessionsRepo{}, &fakeFileStorage{}, attachments)

	view, err := service.ConstructPostPagePostView(post)
	if err != nil {
		t.Fatalf("ConstructPostPagePostView() error = %v", err)
	}
	if view.Post.PostID != 2 || view.Preview.Link == "" {
		t.Fatalf("unexpected post view: %#v", view)
	}
	if len(view.Comments) != 1 || len(view.Comments[0].Attachments) != 1 {
		t.Fatalf("expected one comment with one attachment, got %#v", view.Comments)
	}
}

func TestCreatePostWithoutFiles(t *testing.T) {
	posts := &fakePostsRepo{}
	anons := &fakeAnonsRepo{}
	attachments := &fakeAttachmentsRepo{}
	service := newTestService(posts, &fakeCommentsRepo{}, anons, &fakeSessionsRepo{}, &fakeFileStorage{}, attachments)

	id, err := service.CreatePost(&models.Post{Title: "subject", TextContent: "body"}, map[string]io.Reader{}, &models.Anon{})
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if id != 10 {
		t.Fatalf("post id = %d, want 10", id)
	}
	if posts.withOP.AnonName != "Anon7" || posts.withOP.Avatar != "7" {
		t.Fatalf("unexpected generated op: %#v", posts.withOP)
	}
	if len(attachments.created) != 0 {
		t.Fatalf("created attachments = %d, want 0", len(attachments.created))
	}
}

func TestCreatePostWithFile(t *testing.T) {
	files := &fakeFileStorage{}
	attachments := &fakeAttachmentsRepo{}
	service := newTestService(&fakePostsRepo{}, &fakeCommentsRepo{}, &fakeAnonsRepo{}, &fakeSessionsRepo{}, files, attachments)

	_, err := service.CreatePost(
		&models.Post{Title: "subject", TextContent: "body"},
		map[string]io.Reader{"hello.txt": strings.NewReader("hello")},
		&models.Anon{AnonName: "named"},
	)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if len(files.saved) != 1 {
		t.Fatalf("saved files = %d, want 1", len(files.saved))
	}
	if len(attachments.created) != 1 || attachments.created[0].OriginalName != "hello.txt" {
		t.Fatalf("unexpected attachments: %#v", attachments.created)
	}
}
