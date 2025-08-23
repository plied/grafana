import { DataSourcePluginMeta, DataSourceSettings, LayoutMode } from '@grafana/data';
import { TestingStatus } from '@grafana/runtime';
import { GenericDataSourcePlugin } from 'app/features/datasources/types';

// Extend DataSourceSettings to include allowedTeams field
export interface DataSourceSettingsWithTeams extends DataSourceSettings {
  allowedTeams?: string;
}

export interface DataSourcesState {
  dataSources: DataSourceSettings[];
  searchQuery: string;
  dataSourceTypeSearchQuery: string;
  layoutMode: LayoutMode;
  dataSourcesCount: number;
  dataSource: DataSourceSettingsWithTeams;
  dataSourceMeta: DataSourcePluginMeta;
  isLoadingDataSources: boolean;
  isLoadingDataSourcePlugins: boolean;
  plugins: DataSourcePluginMeta[];
  categories: DataSourcePluginCategory[];
  isSortAscending: boolean;
}

export interface DataSourceSettingsState {
  plugin?: GenericDataSourcePlugin | null;
  testingStatus?: TestingStatus;
  loadError?: string | null;
  loading: boolean;
}

export interface DataSourcePluginCategory {
  id: string;
  title: string;
  plugins: DataSourcePluginMeta[];
}
