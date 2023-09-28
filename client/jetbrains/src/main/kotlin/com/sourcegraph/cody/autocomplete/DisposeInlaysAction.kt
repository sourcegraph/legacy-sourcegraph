package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.editor.actionSystem.EditorAction
import com.intellij.openapi.project.DumbAware

class DisposeInlaysAction : EditorAction(DisposeInlaysActionHandler()), DumbAware {
  init {
    setInjectedContext(true)
  }
}
