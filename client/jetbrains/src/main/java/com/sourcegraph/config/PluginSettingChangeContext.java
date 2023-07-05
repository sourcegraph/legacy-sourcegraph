package com.sourcegraph.config;

import org.jetbrains.annotations.Nullable;

public class PluginSettingChangeContext {
  @Nullable public final String oldUrl;

  @Nullable public final String oldDotComAccessToken;
  @Nullable public final String oldEnterpriseAccessToken;
  public final boolean oldCodyAutoCompleteEnabled;

  @Nullable public final String newUrl;

  @Nullable public final String newDotComAccessToken;
  @Nullable public final String newEnterpriseAccessToken;

  @Nullable public final String newCustomRequestHeaders;
  public final boolean newCodyAutoCompleteEnabled;

  public PluginSettingChangeContext(
      @Nullable String oldUrl,
      @Nullable String oldDotComAccessToken,
      @Nullable String oldEnterpriseAccessToken,
      boolean oldCodyAutoCompleteEnabled,
      @Nullable String newUrl,
      @Nullable String newDotComAccessToken,
      @Nullable String newEnterpriseAccessToken,
      @Nullable String newCustomRequestHeaders,
      boolean newCodyAutoCompleteEnabled) {
    this.oldUrl = oldUrl;
    this.oldDotComAccessToken = oldDotComAccessToken;
    this.oldEnterpriseAccessToken = oldEnterpriseAccessToken;
    this.oldCodyAutoCompleteEnabled = oldCodyAutoCompleteEnabled;
    this.newUrl = newUrl;
    this.newDotComAccessToken = newDotComAccessToken;
    this.newEnterpriseAccessToken = newEnterpriseAccessToken;
    this.newCustomRequestHeaders = newCustomRequestHeaders;
    this.newCodyAutoCompleteEnabled = newCodyAutoCompleteEnabled;
  }
}
