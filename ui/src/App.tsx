import { QueryClient, QueryClientProvider, useMutation, useSuspenseQuery } from "@tanstack/react-query"
import { Suspense } from "react"
import PrimaryButton from "./components/PrimaryButton"
import { RefreshCw, Trash2 } from "lucide-react"
import Card from "./components/Card"

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
  managedTrees: number
  description: string
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
          version: data.version,
          managedTrees: data.total_managed_trees,
          description: data.description
        }))
    }
  })

  const syncMutation = useMutation({
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

  const resetMutation = useMutation({
    mutationFn: async () => {
      return fetch("/api-local/v1/plugin/tbz-baumkataster/reset", {
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
    <main className="flex-1 lg:pl-20">
      <div className="container mt-6">
        <article className="space-y-6 xl:space-y-0 xl:flex xl:items-start xl:space-x-10">
          <div className="xl:w-4/5">
            <h1 className="font-lato font-bold text-3xl mb-4 flex flex-wrap items-center gap-4 lg:text-4xl xl:text-5xl">
              TBZ Baumkataster
            </h1>
            <p className="mb-4">{data.description}</p>
            <div className="flex flex-wrap gap-4 items-center">
            </div>
            <div className="flex flex-wrap gap-4 items-center mb-4">
              <PrimaryButton onClick={() => syncMutation.mutate()}>
                <RefreshCw className={syncMutation.isPending ? "animate-spin" : ""} />
                <span className="font-medium text-base">Sync manuell</span>
              </PrimaryButton >

              <PrimaryButton isDanger onClick={() => resetMutation.mutate()}>
                {resetMutation.isPending ? <RefreshCw className="animate-spin" /> : <Trash2 />}
                <span className="font-medium text-base">Bäume zurücksetzen</span>
              </PrimaryButton>
            </div>
          </div>
        </article>

        <ul className="space-y-5 md:space-y-0 md:grid md:gap-5 md:grid-cols-2 lg:grid-cols-3">
          <li>
            <Card
              overline="Letzte synchronisierung"
              value={data.lastSync.toLocaleString()}
              description="Wann das System sich zuletzt mit dem Baumkataster synchronisiert hat"
            >
            </Card>
          </li>
          <li>
            <Card
              overline="Verwaltete Bäume"
              value={data.managedTrees}
              description="Von diesem Plugin verwaltete Bäume im Green Ecolution System"
            >
            </Card>
          </li>
          <li>
            <Card
              overline="Synchronisierungsintervall"
              value={data.syncInterval}
              description="Ein festgelegtes Intervall, im welche das Baumkataster mit dem System synchronisiert wird"
            >
            </Card>
          </li>
        </ul>



      </div>
    </main>
  )
}

export default App
