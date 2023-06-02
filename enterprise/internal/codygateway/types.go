package codygateway

type EmbeddingsRequest struct {
	// Model is the name of the embeddings model to use.
	Model string `json:"model"`
	// Input is the list of strings to generate embeddings for.
	Input []string `json:"input"`
}

type Embedding struct {
	// Index is the index of the input string this embedding corresponds to.
	Index int `json:"index"`
	// Data is the embedding vector for the input string.
	Data []float32 `json:"data"`
}

type EmbeddingsResponse struct {
	// Embeddings is a list of generated embeddings, one for each input string.
	Embeddings []Embedding `json:"embeddings"`
	// Model is the name of the model used to generate the embeddings.
	Model string `json:"model"`
	// ModelDimensions is the dimensionality of the embeddings model used.
	ModelDimensions int `json:"dimensions"`
}
