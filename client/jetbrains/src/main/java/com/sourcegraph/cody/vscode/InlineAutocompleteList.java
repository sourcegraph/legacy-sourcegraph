package com.sourcegraph.cody.vscode;

import com.sourcegraph.cody.agent.protocol.CompletionEvent;
import org.jetbrains.annotations.Nullable;

import java.util.List;

public class InlineAutocompleteList {
  public final List<InlineAutocompleteItem> items;
  @Nullable public CompletionEvent completionEvent;

  public InlineAutocompleteList(List<InlineAutocompleteItem> items) {
    this.items = items;
  }

  @Override
  public String toString() {
    return "InlineAutocompleteList{" + "items=" + items + '}';
  }
}
