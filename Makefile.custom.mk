# Directories.
SCRIPTS_DIR := hack

sync-crds:
	@echo "$(GEN_COLOR)Syncing Application CRDs with apiextensions$(NO_COLOR)"
	cd $(SCRIPTS_DIR); ./sync-crds.sh
