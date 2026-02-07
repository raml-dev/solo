import { writable } from 'svelte/store';
import { GetConfiguration, UpdateConfiguration } from '../../../wailsjs/go/main/App';
import { configuration } from '../../../wailsjs/go/models';

function createConfigurationStore() {
    const { subscribe, set, update } = writable<configuration.Configuration>(new configuration.Configuration());

    return {
        subscribe,
        init: async () => {
            try {
                const config = await GetConfiguration();
                set(config);
            } catch (error) {
                console.error('Failed to load configuration:', error);
            }
        },
        save: async (newConfig: configuration.Configuration) => {
            try {
                await UpdateConfiguration(newConfig);
                set(newConfig);
            } catch (error) {
                console.error('Failed to save configuration:', error);
                throw error;
            }
        },
        updateSettings: (fn: (cfg: configuration.Configuration) => void) => {
            update(config => {
                const newConfig = new configuration.Configuration(config); // Clone
                fn(newConfig);
                UpdateConfiguration(newConfig).catch(console.error); // Save async
                return newConfig;
            });
        }
    };
}

export const configStore = createConfigurationStore();
