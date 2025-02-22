import { QueryClient, QueryClientProvider, useMutation, useSuspenseQuery } from "@tanstack/react-query"
import { Suspense } from "react"
import PrimaryButton from "./components/PrimaryButton"
import { RefreshCw } from "lucide-react"

const queryClient = new QueryClient()

function App() {
  return (
    <>
      <QueryClientProvider client={queryClient}>
        <div>
          <Suspense fallback={"loading..."}>
            <ControlPanel />
          </Suspense>
        </div>
      </QueryClientProvider>
    </>
  )
}

interface PluginServerInfo {
  syncInterval: string
  lastSync: Date
  slug: string
  version: string
}

const ControlPanel = () => {
  const { data } = useSuspenseQuery<PluginServerInfo>({
    queryKey: ['info'],
    queryFn: async () => {
      return fetch("/api-local/v1/plugin/tbz-baumkataster/info")
        .then(res => {
          if (res.status >= 400) {
            throw res.json()
          }
          return res.json()
        })
        .then(data => ({
          syncInterval: data.sync_interval,
          lastSync: new Date(data.last_sync),
          slug: data.slug,
          version: data.version
        }))
    }
  })

  const mutation = useMutation({
    mutationFn: async () => {
      return fetch("/api-local/v1/plugin/tbz-baumkataster/sync", {
        method: "POST"
      }).then(res => {
        if (res.status >= 400) {
          throw res.json()
        }
        return
      })
    },
    onSuccess: () => console.log("done!"),
    onError: (error, variables, context) => console.error("error", error, variables, context),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['info'] })
    },
  })

  return (
    <>
      <PrimaryButton onClick={() => mutation.mutate()}>
        <RefreshCw className={mutation.isPending ? "animate-spin" : ""} />
        <span className="font-medium text-base">Sync manuell</span>

      </PrimaryButton>

      <p>Last Sync: {data.lastSync.toLocaleString()}</p>
    </>
  )
}

export default App
