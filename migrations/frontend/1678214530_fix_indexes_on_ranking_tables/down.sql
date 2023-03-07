DROP INDEX IF EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id;
DROP INDEX IF EXISTS codeintel_ranking_references_graph_key_id;
DROP INDEX IF EXISTS codeintel_ranking_definitions_graph_key_symbol_search;
DROP INDEX IF EXISTS codeintel_path_ranks_graph_key;
DROP INDEX IF EXISTS codeintel_path_ranks_repository_id_updated_at_id;

CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_and_repository_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_path_counts_inputs_graph_key_repository_id_id ON codeintel_ranking_path_counts_inputs(graph_key, repository_id, id) INCLUDE (document_path) WHERE NOT processed;
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_upload_id ON codeintel_ranking_definitions(upload_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_symbol_name ON codeintel_ranking_definitions(symbol_name);
CREATE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks(repository_id);
CREATE INDEX IF NOT EXISTS codeintel_path_ranks_updated_at ON codeintel_path_ranks(updated_at);
