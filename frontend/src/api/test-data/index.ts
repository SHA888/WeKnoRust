import { get, setTestData } from '../../utils/request';

export interface TestDataResponse {
  success: boolean;
  data: {
    tenant: {
      id: number;
      name: string;
      api_key: string;
    };
    knowledge_bases: Array<{
      id: string;
      name: string;
      description: string;
    }>;
  }
}

// Whether test data has been loaded
let isTestDataLoaded = false;

/**
 * Load test data
 * Call this before API calls to ensure test data is loaded
 * @returns Promise<boolean> whether loading succeeded
 */
export async function loadTestData(): Promise<boolean> {
  // If already loaded, return early
  if (isTestDataLoaded) {
    return true;
  }

  try {
    console.log('Start loading test data...');
    const response = await get('/api/v1/test-data');
    console.log('Test data response', response);
    
    if (response && response.data) {
      // Set test data
      setTestData({
        tenant: response.data.tenant,
        knowledge_bases: response.data.knowledge_bases
      });
      isTestDataLoaded = true;
      console.log('Test data loaded successfully');
      return true;
    } else {
      console.warn('Test data response is empty');
      return false;
    }
  } catch (error) {
    console.error('Failed to load test data:', error);
    return false;
  }
} 
